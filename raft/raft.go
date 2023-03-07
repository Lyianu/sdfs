package raft

import (
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	FOLLOWER = iota
	LEAD
	CANDIDATE
)

var maxRTT = 150

type ConsensusModule struct {
	id      int
	peerIds []int
	server  *Server

	currentTerm int
	votedFor    int

	state int

	electionResetEvent time.Time

	mu sync.Mutex
}

// NewConsensusModule creates a ConsensusModule, it takes an address of
// arbitrary Master server to get peerIds, when passed with empty string
// it creates a server on its own
func NewConsensusModule(serverAddr string) *ConsensusModule {
	cm := &ConsensusModule{
		id:      rand.Int(),
		peerIds: []int{},
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
			// cm.startElection()
			cm.mu.Unlock()
			return
		}
		cm.mu.Unlock()
	}
}

func (cm *ConsensusModule) startElection() {
	cm.state = CANDIDATE
	cm.currentTerm += 1
	// record term when the election start
	savedCurrentTerm := cm.currentTerm
	cm.electionResetEvent = time.Now()
	cm.votedFor = cm.id

	voteReceived := 1

	for _, peerId := range cm.peerIds {
		go func(peerId int) {
			args := RequestVoteRequest{
				Term:        uint64(savedCurrentTerm),
				CandidateId: uint64(cm.id),
			}
			//var reply RequestVoteResponse

			if reply, err := cm.server.RequestVote(args); err != nil {
				cm.mu.Lock()
				defer cm.mu.Unlock()

				if cm.state != CANDIDATE {
					// not in CANDIDATE state
					return
				}

				if reply.Term > savedCurrentTerm {
					// term expired
					// cm.becomeFollower(reply.Term)
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
			}
		}()
	}
}

func (cm *ConsensusModule) electionTimeout() time.Duration {
	if len(os.Getenv("RAFT_FORCE_MORE_REELECTION")) > 0 && rand.Intn(3) == 0 {
		return time.Duration(maxRTT) * time.Millisecond
	} else {
		return time.Duration(maxRTT+rand.Intn(maxRTT)) * time.Millisecond
	}
}
