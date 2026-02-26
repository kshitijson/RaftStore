package raft

import (
	"fmt"
	"strings"
	"time"
)

func (n *Node) applyEntries() {

	for {

		n.Mu.Lock()

		for n.LastApplied < n.CommitIndex {
			n.LastApplied++

			entry := n.Log[n.LastApplied]
			n.applyCommand(entry.Command)
			if len(n.Log) > 100 {
				n.CreateSnapshot()
			}
		}

		n.Mu.Unlock()

		time.Sleep(time.Millisecond)

	}


}

func (n *Node) applyCommand(cmd []byte) {

	parts := strings.Split(string(cmd), " ")

	if len(parts) != 3 {
		return
	}

	if parts[0] == "PUT" {

		key := parts[1]
		value := parts[2]

		n.KV.Mu.Lock()
		n.KV.Data[key] = value
		n.KV.Mu.Unlock()

		fmt.Printf("Node %d applied: %s=%s\n", n.Id, key, value)

	}

}