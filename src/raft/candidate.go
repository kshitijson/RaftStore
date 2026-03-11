package raft

import (
	"fmt"
)

func (n *Node) runCandidate() {

	// Fresh timer so this candidate doesn't immediately time out.
	n.ResetElectionTimer()

	// startElection increments the term, self-votes, broadcasts RequestVote RPCs,
	// and returns the initial vote count (1) and the voter set (self).
	startElection := func() (int, map[int]bool) {
		n.Mu.Lock()
		n.CurrentTerm++ // new term for this election round
		n.VotedFor = n.Id // vote for self
		n.Mu.Unlock()

		fmt.Printf("Node %d became CANDIDATE (term %d)\n", n.Id, n.CurrentTerm)

		// Send RequestVote RPCs to all peers in parallel.
		n.broadcastVoteRequestRPC()

		// Start with 1 vote (self) and mark self as already voted.
		return 1, map[int]bool{n.Id: true}
	}

	votes, voters := startElection()


	for {
		select {

		// A vote response (or other inbox message) arrived.
		case msg := <-n.Inbox:
			// Returns true if the candidate should exit (won or stepped down).
			if n.handleCandidateVoteResponse(msg, &votes, voters) {
				return
			}

		// Election timed out — no majority reached, restart with a new term.
		case <-n.ElectionTimer.C:
			fmt.Printf("Node %d restarted election (term %d)\n", n.Id, n.CurrentTerm+1)

			// Drain the timer channel if it fired between Stop and here.
			if !n.ElectionTimer.Stop() {
				<-n.ElectionTimer.C
			}

			votes, voters = startElection()
			n.ResetElectionTimer()

		// Another goroutine signalled a state change (e.g. stepped down), exit.
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

	// If we've already transitioned away from Candidate, discard the message.
	if n.State != Candidate {
		n.Mu.Unlock()
		return true
	}

	// Ignore non-VoteResponse messages (e.g. AppendEntries that slipped in).
	resp, ok := msg.(VoteResponse)
	if !ok {
		n.Mu.Unlock()
		return false
	}

	// Peer has a higher term — we're outdated, step down to follower.
	// Must unlock before StepDown since it acquires the lock internally.
	if resp.Term > n.CurrentTerm {
		n.Mu.Unlock()
		n.StepDown(resp.Term)
		return true
	}

	// Discard: vote is from a past term, was denied, or is a duplicate.
	if resp.Term < n.CurrentTerm || !resp.Granted || voters[resp.From] {
		n.Mu.Unlock()
		return false
	}

	// Record this new granted vote.
	voters[resp.From] = true
	(*votes)++

	// +1 to include self; win if we have strictly more than half.
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