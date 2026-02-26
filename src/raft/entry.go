package raft

import "fmt"

func (n *Node) Run() {

	fmt.Println("Entered in Run Loop")

	go n.applyEntries()

	for {
		switch n.State {
		case Follower:
			n.runFollower()
		case Candidate:
			n.runCandidate()
		case Leader:
			n.runLeader()
		}

	}

}