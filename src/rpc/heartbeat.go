package rpc

import (
	"context"
	"fmt"

	"github.com/etcd/raft/raftpb"
	"github.com/etcd/src/raft"
)

func (s *RaftServer) AppendEntries(
	ctx context.Context,
	req *raftpb.AppendEntriesRequest,
) (*raftpb.AppendEntriesResponse, error) {

	fmt.Printf("Node %d received heartbeat from %d (term %d)\n", s.node.Id, req.LeaderId, req.Term)

	n := s.node

	if int(req.Term) < n.CurrentTerm {
		return &raftpb.AppendEntriesResponse{
			Term: int32(n.CurrentTerm),
			Success: false,
		}, nil
	}

	if int(req.Term) > n.CurrentTerm {
		n.StepDown(int(req.Term))
	}

	// Update Leader Info
	n.Mu.Lock()
	n.LeaderId = int(req.LeaderId)
	n.LeaderAddr = n.PeerAddrMap[n.LeaderId]
	n.Mu.Unlock()

	fmt.Printf("Leader address: %s\n", n.LeaderAddr)

	n.ResetElectionTimer()

	// Log inconsistency Check
	if int(req.PrevLogIndex) > 0 {
		if int(req.PrevLogIndex) >= len(n.Log) || n.Log[req.PrevLogIndex].Term != int(req.PrevLogTerm) {
			return &raftpb.AppendEntriesResponse{
				Term: int32(n.CurrentTerm),
				Success: false,
			}, nil

		}
	}

	// Append new Entries
	n.Mu.Lock()
	for _, entry := range req.Entries {

		index := int(entry.Index)

		if index < len(n.Log) {
			if n.Log[index].Term != int(entry.Term) {
				n.Log = n.Log[:index]
			}
		}

		if index >= len(n.Log) {
			n.Log = append(n.Log, raft.LogEntry{
				Term: int(entry.Term),
				Index: index,
				Command: entry.Command,
			})
		}

	}

	if int(req.LeaderCommit) > n.CommitIndex {
		n.CommitIndex = min(int(req.LeaderCommit), len(n.Log)-1)
	}
	n.Mu.Unlock()

	n.Persist()

	return &raftpb.AppendEntriesResponse{
		Term: int32(n.CurrentTerm),
		Success: true,
	}, nil
}

