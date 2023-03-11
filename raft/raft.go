package raft

import (
	"context"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/Lyianu/sdfs/log"
)

type CMState int

const (
	FOLLOWER CMState = iota
	LEADER
	CANDIDATE
	DEAD
)

func (s CMState) String() string {
	switch s {
	case FOLLOWER:
		return "FOLLOWER"
	case LEADER:
		return "LEADER"
	case CANDIDATE:
		return "CANDIDATE"
	case DEAD:
		return "DEAD"
	default:
		return "UNREACHABLE"
	}
}

var maxRTT = 150

type LogEntry struct {
	Command interface{}
	Term    uint64
}

type CommitEntry struct {
	Command interface{}
	Index   uint64
	Term    uint64
}

type ConsensusModule struct {
	id         int32
	peerIds    []int32
	server     *Server
	nextIndex  map[int32]uint64
	matchIndex map[int32]uint64

	commitChan         chan<- CommitEntry
	newCommitReadyChan chan struct{}

	currentTerm   uint64
	votedFor      int32
	log           []LogEntry
	currentLeader int32

	commitIndex        uint64
	state              CMState
	electionResetEvent time.Time

	mu sync.Mutex
}

func init() {
	rand.Seed(time.Now().Unix())
}

// NewConsensusModule creates a ConsensusModule, its peer list will be set by
// server at first
func NewConsensusModule(ready <-chan struct{}) *ConsensusModule {
	cm := &ConsensusModule{
		id:          rand.Int31(),
		nextIndex:   make(map[int32]uint64),
		matchIndex:  make(map[int32]uint64),
		commitIndex: 0,
		votedFor:    -1,
		currentTerm: 0,
	}

	go func() {
		<-ready
		cm.mu.Lock()
		cm.electionResetEvent = time.Now()
		cm.mu.Unlock()
		cm.runElectionTimer()
	}()
	return cm
}

// Submit tries to append entry to the log, it returns leader id on failure
func (cm *ConsensusModule) Submit(cmd interface{}) (bool, int32) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	log.Debugf("Submit received by %v: %v", cm.state, cmd)
	if cm.state == LEADER {
		log.Debugf("log=%v", cm.log)
		cm.log = append(cm.log, LogEntry{Command: cmd, Term: cm.currentTerm})
		return true, cm.id
	}
	log.Debugf("cm is not leader, leader ID: %d", cm.currentLeader)
	return false, cm.currentLeader
}

func (cm *ConsensusModule) runElectionTimer() {
	timeoutDuration := cm.electionTimeout()
	cm.mu.Lock()
	termStarted := cm.currentTerm
	cm.mu.Unlock()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		<-ticker.C

		cm.mu.Lock()
		if cm.state != CANDIDATE && cm.state != FOLLOWER {
			// become LEADER
			cm.mu.Unlock()
			return
		}

		if termStarted != cm.currentTerm {
			// Term changed
			cm.mu.Unlock()
			return
		}

		if elapsed := time.Since(cm.electionResetEvent); elapsed >= timeoutDuration {
			// timeout occurs, become Candidate
			cm.mu.Unlock()
			cm.startElection()
			return
		}
		cm.mu.Unlock()
	}
}

func (cm *ConsensusModule) startElection() {
	log.Infof("Election started, term: %d, id: %d", cm.currentTerm+1, cm.id)
	cm.state = CANDIDATE
	cm.currentTerm += 1
	// record term when the election start
	savedCurrentTerm := cm.currentTerm
	cm.electionResetEvent = time.Now()
	cm.votedFor = cm.id

	voteReceived := 1
	if len(cm.peerIds) == 0 {
		log.Debugf("no peers, elect self as leader")
		cm.startLeader()
		return
	}

	for _, peerId := range cm.peerIds {
		go func(peerId int32) {
			log.Debugf("[CLIENT]RequestVote(%d)\n", peerId)
			args := RequestVoteRequest{
				Term:        uint64(savedCurrentTerm),
				CandidateId: int32(cm.id),
			}
			//var reply RequestVoteResponse
			for k, v := range cm.server.peers {
				log.Debugf("cm.server.peers: %d, %+v", k, v)

			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			if reply, err := cm.server.peers[peerId].RequestVote(ctx, &args); err == nil {
				log.Debugf("RequestVote(%d), resp: %+v", peerId, *reply)
				cm.mu.Lock()

				if cm.state != CANDIDATE {
					// not in CANDIDATE state
					cm.mu.Unlock()
					return
				}

				if reply.Term > savedCurrentTerm {
					// term expired
					cm.becomeFollower(reply.Term, peerId)
					cm.mu.Unlock()
					return
				} else if reply.Term == savedCurrentTerm {
					if reply.VoteGranted {
						voteReceived++
						// enough votes, win the election
						if voteReceived*2 > len(cm.peerIds)+1 {
							cm.mu.Unlock()
							cm.startLeader()
							return
						}
					}
					cm.mu.Unlock()
				}
			} else {
				log.Errorf("requestVote: error: %q", err)
			}
		}(peerId)
	}
	log.Infof("election RV queue ended(might have unfinished RVs)")
}

func (cm *ConsensusModule) electionTimeout() time.Duration {
	if len(os.Getenv("RAFT_FORCE_MORE_REELECTION")) > 0 && rand.Intn(3) == 0 {
		return time.Duration(maxRTT) * time.Millisecond
	} else {
		return time.Duration(8*maxRTT+rand.Intn(maxRTT)) * time.Millisecond
	}
}

func (cm *ConsensusModule) becomeFollower(term uint64, leader int32) {
	log.Infof("FOLLOWER started, term: %d, id: %d", term, cm.id)
	cm.state = FOLLOWER
	cm.currentTerm = term
	cm.currentLeader = leader
	cm.votedFor = -1
	cm.electionResetEvent = time.Now()

	go cm.runElectionTimer()
}

func (cm *ConsensusModule) startLeader() {
	log.Infof("LEADER started, term: %d, id: %d", cm.currentTerm, cm.id)
	cm.state = LEADER
	cm.mu.Lock()
	cm.currentLeader = cm.id
	cm.mu.Unlock()

	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			cm.leaderSendHeartbeats()
			<-ticker.C

			cm.mu.Lock()
			if cm.state != LEADER {
				cm.mu.Unlock()
				return
			}
			cm.mu.Unlock()
		}
	}()
}

func (cm *ConsensusModule) leaderSendHeartbeats() {
	log.Debugf("Sending HB...\n")
	cm.mu.Lock()
	if cm.state != LEADER {
		cm.mu.Unlock()
		return
	}
	savedCurrentTerm := cm.currentTerm
	cm.mu.Unlock()

	for _, peerId := range cm.peerIds {
		go func(peerId int32) {
			cm.mu.Lock()
			ni, ok := cm.nextIndex[peerId]
			if !ok {
				ni = 0
				cm.nextIndex[peerId] = 0
				cm.matchIndex[peerId] = 0
			}
			prevLogIndex := int64(ni) - 1
			prevLogTerm := -1
			if prevLogIndex >= 0 {
				prevLogTerm = int(cm.log[prevLogIndex].Term)
			}
			entries := cm.log[ni:]
			entry := []*Entry{}
			for _, v := range entries {
				entry = append(entry, &Entry{Term: v.Term, Data: []byte(v.Command.([]byte))})
			}

			log.Debugf("Sending HB to %d\n", peerId)
			req := AppendEntriesRequest{
				Term:         savedCurrentTerm,
				LeaderId:     cm.id,
				PrevLogIndex: uint64(prevLogIndex),
				PrevLogTerm:  uint64(prevLogTerm),
				Entries:      entry,
				LeaderCommit: cm.commitIndex,
			}
			cm.mu.Unlock()
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			resp, err := cm.server.peers[peerId].AppendEntries(ctx, &req)
			if err == nil {
				cm.mu.Lock()
				defer cm.mu.Unlock()
				if resp.Term > savedCurrentTerm {
					cm.becomeFollower(resp.Term, resp.LeaderId)
					return
				}

				if cm.state == LEADER && savedCurrentTerm == resp.Term {
					if resp.Success {
						cm.nextIndex[peerId] = ni + uint64(len(entry))
						cm.matchIndex[peerId] = cm.nextIndex[peerId] - 1
						log.Debugf("AE resp from %d success: nextIndex := %v, matchIndex := %v", peerId, cm.nextIndex, cm.matchIndex)

						savedCommitIndex := cm.commitIndex
						for i := cm.commitIndex + 1; i < uint64(len(cm.log)); i++ {
							matchCount := 1
							for _, peerId := range cm.peerIds {
								if cm.matchIndex[peerId] >= i {
									matchCount++
								}
							}
							if matchCount*2 > len(cm.peerIds)+1 {
								cm.commitIndex = 1
							}
						}
						if cm.commitIndex != savedCommitIndex {
							cm.newCommitReadyChan <- struct{}{}
						}
					} else {
						cm.nextIndex[peerId] = ni - 1
					}
				}
			}
		}(peerId)
	}
}

func (cm *ConsensusModule) AppendEntries(req *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	log.Debugf("AppendEntries: %+v", *req)

	if req.Term > cm.currentTerm {
		log.Debug(" - term out of data in AE")
		cm.becomeFollower(req.Term, req.LeaderId)
	}

	resp := &AppendEntriesResponse{
		Success: false,
	}
	if req.Term == cm.currentTerm {
		// only LEADER sends AEs
		if cm.state != FOLLOWER {
			cm.becomeFollower(req.Term, req.LeaderId)
		}
		cm.electionResetEvent = time.Now()
		resp.Success = true
	}

	if req.Term < cm.currentTerm {
		resp.LeaderId = cm.currentLeader
	}

	resp.Term = cm.currentTerm
	log.Debugf("AppendEntries resp: %+v", *resp)
	return resp, nil
}
