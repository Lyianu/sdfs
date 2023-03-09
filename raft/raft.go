package raft

import (
	"context"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/Lyianu/sdfs/log"
)

const (
	FOLLOWER = iota
	LEADER
	CANDIDATE
)

var maxRTT = 150

type LogEntry struct {
	Command interface{}
	Term    int
}

type ConsensusModule struct {
	id      int32
	peerIds []int32
	server  *Server

	currentTerm uint64
	votedFor    int32
	log         []LogEntry

	state              int
	electionResetEvent time.Time

	mu sync.Mutex
}

func init() {
	rand.Seed(time.Now().Unix())
}

// NewConsensusModule creates a ConsensusModule, its peer list will be set by
// server at first
func NewConsensusModule() *ConsensusModule {
	cm := &ConsensusModule{
		id:          rand.Int31(),
		currentTerm: 0,
	}
	return cm
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
			cm.startElection()
			cm.mu.Unlock()
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
				log.Debugf("RequestVote(%d), resp: %+v", reply.Term)
				cm.mu.Lock()
				defer cm.mu.Unlock()

				if cm.state != CANDIDATE {
					// not in CANDIDATE state
					return
				}

				if reply.Term > savedCurrentTerm {
					// term expired
					cm.becomeFollower(reply.Term)
					return
				} else if reply.Term == savedCurrentTerm {
					if reply.VoteGranted {
						voteReceived++
						// enough votes, win the election
						if voteReceived*2 > len(cm.peerIds)+1 {
							cm.startLeader()
							return
						}
					}
				}
			} else {
				log.Errorf("requestVote: error: %q", err)
			}
		}(peerId)
	}
}

func (cm *ConsensusModule) electionTimeout() time.Duration {
	if len(os.Getenv("RAFT_FORCE_MORE_REELECTION")) > 0 && rand.Intn(3) == 0 {
		return time.Duration(maxRTT) * time.Millisecond
	} else {
		return time.Duration(8*maxRTT+rand.Intn(maxRTT)) * time.Millisecond
	}
}

func (cm *ConsensusModule) becomeFollower(term uint64) {
	log.Infof("FOLLOWER started, term: %d, id: %d", term, cm.id)
	cm.state = FOLLOWER
	cm.currentTerm = term
	cm.votedFor = -1
	cm.electionResetEvent = time.Now()

	go cm.runElectionTimer()
}

func (cm *ConsensusModule) startLeader() {
	log.Infof("LEADER started, term: %d, id: %d", cm.currentTerm, cm.id)
	cm.state = LEADER

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
			log.Debugf("Sending HB to %d\n", peerId)
			req := AppendEntriesRequest{Term: savedCurrentTerm, LeaderId: cm.id}
			ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
			defer cancel()
			resp, err := cm.server.peers[peerId].AppendEntries(ctx, &req)
			if err == nil {
				cm.mu.Lock()
				defer cm.mu.Unlock()
				if resp.Term > savedCurrentTerm {
					cm.becomeFollower(resp.Term)
					return
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
		cm.becomeFollower(req.Term)
	}

	resp := &AppendEntriesResponse{
		Success: false,
	}
	if req.Term == cm.currentTerm {
		// only LEADER sends AEs
		if cm.state != FOLLOWER {
			cm.becomeFollower(req.Term)
		}
		cm.electionResetEvent = time.Now()
		resp.Success = true
	}

	resp.Term = cm.currentTerm
	log.Debugf("AppendEntries resp: %+v", *resp)
	return resp, nil
}
