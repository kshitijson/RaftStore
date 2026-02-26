package raft

import "time"

func (n *Node) runFollower() {

	timer := time.NewTimer(randomElectionTimeout())
	defer timer.Stop()

	// expose timer so RPC handlers can reset it
	n.ElectionTimer = timer

	for {

		if n.State != Follower {
			return
		}

		select {
		case <-timer.C:
			n.Mu.Lock()
			n.ResetElectionTimer()
			n.State = Candidate
			n.Mu.Unlock()
			return

		case <-n.StateCh:
			return
		}

	}

}	
