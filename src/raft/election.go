package raft

import (
	"math/rand"
	"time"
)

// 150ms ≤ duration ≤ 299ms
func randomElectionTimeout() time.Duration {
	return time.Duration(1500 + rand.Intn(1500)) * time.Millisecond
}

func (n *Node) ResetElectionTimer() {

	if n.ElectionTimer == nil {
		n.ElectionTimer = time.NewTimer(randomElectionTimeout())
		return
	}

	if !n.ElectionTimer.Stop() {
		select {
		case <-n.ElectionTimer.C:
		default:
		}
	}

	n.ElectionTimer.Reset(randomElectionTimeout())
}