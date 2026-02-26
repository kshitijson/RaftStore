package raft

import (
	"context"
	"time"

	"github.com/etcd/raft/raftpb"
)	

func (n *Node) replicateLogs() {

	n.Mu.Lock()
	if n.State != Leader {
		n.Mu.Unlock()
		return
	}
	currentTerm := n.CurrentTerm
	n.Mu.Unlock()

	for peerID, client := range n.Clients {
		go n.replicateToPeer(peerID, client, currentTerm)
	}

}

func (n *Node) replicateToPeer(
	id int,
	c raftpb.RaftServiceClient,
	currentTerm int,
) {

	req := n.buildAppendEntriesRequest(id, currentTerm)
	if req == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := c.AppendEntries(ctx, req)
	if err != nil {
		return
	}

	n.handleAppendResponse(id, resp)	

}

func (n *Node) buildAppendEntriesRequest(
	id int,
	currentTerm int,
) *raftpb.AppendEntriesRequest {

	n.Mu.Lock()	
	defer n.Mu.Unlock()

	nextIndex := n.NextIndex[id]
	prevLogIndex := nextIndex - 1

	prevLogTerm := n.getPrevLogTerm(prevLogIndex)
	entries := n.getEntriesFrom(nextIndex)

	return &raftpb.AppendEntriesRequest{
		Term: int32(currentTerm),
		LeaderId: int32(n.Id),
		PrevLogIndex: int32(prevLogIndex),
		PrevLogTerm: prevLogTerm,
		Entries: entries,
		LeaderCommit: int32(n.CommitIndex),
	}
	
}	

func (n *Node) getPrevLogTerm(prevLogIndex int) int32 {
	if prevLogIndex <= 0 || prevLogIndex >= len(n.Log) {
		return 0
	}
	return int32(n.Log[prevLogIndex].Term)
}

func (n *Node) getEntriesFrom(nextIndex int) []*raftpb.LogEntry {

	if nextIndex >= len(n.Log) {
		return nil
	}

	entries := make([]*raftpb.LogEntry, 0, len(n.Log[nextIndex:]))

	for _, e := range n.Log[nextIndex:] {
		entries = append(entries, &raftpb.LogEntry{
			Term: int32(e.Term),
			Index: int32(e.Index),
			Command: e.Command,
		})
	}

	return entries
}

func (n *Node) handleAppendResponse(peerID int, resp *raftpb.AppendEntriesResponse) {

	n.Mu.Lock()

	if int(resp.Term) > n.CurrentTerm {
		n.Mu.Unlock()
		n.StepDown(int(resp.Term))
		return
	}

	if resp.Success {
		n.MatchIndex[peerID] = n.NextIndex[peerID]
		n.NextIndex[peerID] = n.MatchIndex[peerID] + 1
		n.updateCommitIndex()
	} else {
		if n.NextIndex[peerID] > 1 {
			n.NextIndex[peerID]--
		}
	}

	n.Mu.Unlock()

}