package raft

import (
	"fmt"
	"strings"
	"time"
)

// applyEntries runs as a background goroutine and continuously catches up
// LastApplied to CommitIndex. CommitIndex is advanced by the leader once a
// majority of nodes have stored the entry; LastApplied tracks how far the
// state machine has actually executed. The two indices move independently —
// this loop is the bridge between them.
func (n *Node) applyEntries() {

	for {

		n.Mu.Lock()

		// Apply every committed entry that hasn't been executed yet.
		for n.LastApplied < n.CommitIndex {
			n.LastApplied++

			entry := n.Log[n.LastApplied]
			n.applyCommand(entry.Command)

			// Compact the log once it grows too large to keep memory bounded.
			if len(n.Log) > 100 {
				n.CreateSnapshot()
			}
		}

		n.Mu.Unlock()

		// Poll every 1ms rather than blocking on a channel/condition variable.
		time.Sleep(time.Millisecond)

	}


}

// applyCommand deserializes a raw log entry and executes it against the KV store.
// Commands are space-separated strings with the format: "OP key value".
// Any entry that doesn't match this 3-part format (e.g. the dummy "init" entry)
// is silently ignored.
func (n *Node) applyCommand(cmd []byte) {

	parts := strings.Split(string(cmd), " ")

	if len(parts) != 3 {
		return
	}

	if parts[0] == "PUT" {

		key := parts[1]
		value := parts[2]

		// Use the KV store's own mutex here (not the node mutex) since the
		// data store and Raft state are protected independently.
		n.KV.Mu.Lock()
		n.KV.Data[key] = value
		n.KV.Mu.Unlock()

		fmt.Printf("Node %d applied: %s=%s\n", n.Id, key, value)

	}

}