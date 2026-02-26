package raft

import "fmt"

func (n *Node) HandleVoteRequestRPC(rv RequestVote) bool {

	fmt.Printf("Node %d received vote request from %d (term %d)\n", n.Id, rv.From, rv.Term)

	n.Mu.Lock()
	if rv.Term < n.CurrentTerm {
		n.Mu.Unlock()
		return false
	}

	if rv.Term > n.CurrentTerm {
		n.Mu.Unlock()
		n.StepDown(rv.Term)
		n.Mu.Lock()
	}

	if n.VotedFor == -1 || n.VotedFor == rv.From {

		// Log completness check
		myLastIndex := len(n.Log) - 1
		myLastTerm := n.Log[myLastIndex].Term

		logOk := rv.LastLogTerm > myLastTerm || (rv.LastLogTerm == myLastTerm && rv.LastLogIndex >= myLastIndex)

		if !logOk {
			n.Mu.Unlock()
			return false // candidate's log is stale deny vote
		}

		n.VotedFor = rv.From
		n.ResetElectionTimer()
		n.Mu.Unlock()
		return true
	}

	n.Mu.Unlock()
	return false

}