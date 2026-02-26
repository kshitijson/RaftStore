package raft

func (n *Node) updateCommitIndex() {

	for i := len(n.Log) - 1; i > n.CommitIndex; i-- {

		count := 1
		
		for _, matchIndex := range n.MatchIndex {
			if matchIndex >= i {
				count++
			}
		}

		if count > (len(n.Clients)+1)/2 && n.Log[i].Term == n.CurrentTerm {
			n.CommitIndex = i
			break
		}

	}

}

func (n *Node) StepDown(term int) {

	n.Mu.Lock()
	defer n.Mu.Unlock()

	if term > n.CurrentTerm {
		n.CurrentTerm = term
	}

	n.State = Follower
	n.VotedFor = -1;

	n.Persist()

	// notify running loop to exit
	select {
		case n.StateCh <- struct{}{}:
		default:
	}
}