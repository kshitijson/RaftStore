package raft

import (
	"fmt"
)

func (n *Node) runCandidate() {

	n.ResetElectionTimer()

	startElection := func() (int, map[int]bool) {
		n.Mu.Lock()
		n.CurrentTerm++
		n.VotedFor = n.Id
		n.Mu.Unlock()

		fmt.Printf("Node %d became CANDIDATE (term %d)\n", n.Id, n.CurrentTerm)

		n.broadcastVoteRequestRPC()

		return 1, map[int]bool{n.Id: true}
	}

	votes, voters := startElection()


	for {
		select {

		case msg := <-n.Inbox:
			if n.handleCandidateVoteResponse(msg, &votes, voters) {
				return
			}

		case <-n.ElectionTimer.C:
			fmt.Printf("Node %d restarted election (term %d)\n", n.Id, n.CurrentTerm+1)

			if !n.ElectionTimer.Stop() {
				<-n.ElectionTimer.C
			}

			votes, voters = startElection()
			n.ResetElectionTimer()			

		case <-n.StateCh:
			return
		}
	}


}

func (n *Node) handleCandidateVoteResponse(
	msg interface{},
	votes *int,
	voters map[int]bool,
) bool {

	n.Mu.Lock()

	if n.State != Candidate {
		n.Mu.Unlock()
		return true
	}

	resp, ok := msg.(VoteResponse)
	if !ok {
		n.Mu.Unlock()
		return false
	}

	// Higher term -> step down (must unlock before calling StepDown)
	if resp.Term > n.CurrentTerm {
		n.Mu.Unlock()
		n.StepDown(resp.Term)
		return true
	}

	// Ignore stale || rejected || duplicate votes
	if resp.Term < n.CurrentTerm || !resp.Granted || voters[resp.From] {
		n.Mu.Unlock()
		return false
	}

	voters[resp.From] = true
	(*votes)++

	// Majority Reached
	totalNodes := len(n.Clients) + 1
	if *votes > totalNodes/2 {
		fmt.Printf("Node %d became LEADER (term %d)\n", n.Id, n.CurrentTerm)
		n.State = Leader
		n.Mu.Unlock()
		n.ResetElectionTimer()
		return true
	}

	n.Mu.Unlock()
	return false
}