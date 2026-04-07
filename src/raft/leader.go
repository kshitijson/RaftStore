package raft

import (
	"fmt"
	"time"
)


func (n *Node) runLeader() {

	n.Mu.Lock()

	fmt.Printf("Node %d became LEADER (term %d)\n", n.Id, n.CurrentTerm)

	// Advertise self as leader so followers can redirect client requests here.
	n.LeaderId = n.Id
	n.LeaderAddr = n.SelfAddr

	// Fresh maps each term — stale NextIndex/MatchIndex from a previous leadership
	// would cause incorrect log replication decisions.
	n.NextIndex = make(map[int]int)
	n.MatchIndex = make(map[int]int)

	// Append a no-op entry so the leader can commit entries from previous terms.
	// Without this, a leader may never advance CommitIndex if no new client writes arrive,
	// leaving old entries in limbo (see Raft paper §8).
	n.Log = append(n.Log, LogEntry{
		Term:    n.CurrentTerm,
		Index:   len(n.Log),
		Command: []byte("noop"),
	})

	lastIndex := len(n.Log)

	for peerID := range n.Clients {
		// Optimistically assume each peer is fully caught up; replication will
		// decrement NextIndex if a follower's log is actually behind.
		n.NextIndex[peerID] = lastIndex
		// MatchIndex starts at 0 (unknown) and advances as followers confirm entries.
		n.MatchIndex[peerID] = 0
	}

	n.Mu.Unlock()

	// Send heartbeats and replicate logs every 100ms — well below the election
	// timeout so followers never falsely trigger an election while a leader is alive.
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {

		// Guard against the case where State changed between a ticker fire and
		// this check (e.g. a higher-term RPC caused a stepdown).
		if n.State != Leader {
			return
		}

		select {
		// On each tick, replicate any new log entries and send heartbeats to all peers.
		case <-ticker.C:
			n.replicateLogs()

		// Another goroutine signalled a state change (e.g. stepped down), exit.
		case <-n.StateCh:
			return
		}
	}

}