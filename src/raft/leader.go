package raft

import (
	"fmt"
	"time"
)


func (n *Node) runLeader() {

	n.Mu.Lock()

	fmt.Printf("Node %d became LEADER (term %d)\n", n.Id, n.CurrentTerm)

	n.LeaderId = n.Id
	n.LeaderAddr = n.SelfAddr

	n.NextIndex = make(map[int]int)
	n.MatchIndex = make(map[int]int)

	n.Log = append(n.Log, LogEntry{
		Term: n.CurrentTerm,
		Index: len(n.Log),
		Command: []byte("noop"),
	})

	lastIndex := len(n.Log)

	for peerID := range n.Clients {
		n.NextIndex[peerID] = lastIndex
		n.MatchIndex[peerID] = 0
	}

	n.Mu.Unlock()


	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {

		if n.State != Leader {
			return
		}

		select {
		case <-ticker.C:
			n.replicateLogs()

		case <-n.StateCh:
			return
		}
	}

}