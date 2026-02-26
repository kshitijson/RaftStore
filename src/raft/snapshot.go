package raft

func (n *Node) CreateSnapshot() {

	snapshot := Snapshot{
		LastIncludedIndex: n.CommitIndex,
		LastIncludedTerm: n.CurrentTerm,
		KVData: n.KV.Data,
	}

	n.Storage.SaveSnapshot(snapshot)

	// truncate log
	n.Log = n.Log[n.CommitIndex:]

}