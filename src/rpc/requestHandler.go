package rpc

import (
	"context"

	"github.com/etcd/raft/raftpb"
	"github.com/etcd/src/raft"
)

func (s *RaftServer) RequestVote(
	ctx context.Context,
	req *raftpb.RequestVoteRequest,
) (*raftpb.RequestVoteResponse, error) {

	rv := raft.RequestVote{
		From: int(req.CandidateId),
		Term: int(req.Term),
		LastLogIndex: int(req.LastLogIndex),
		LastLogTerm: int(req.LastLogTerm),
	}

	granted := s.node.HandleVoteRequestRPC(rv)

	s.node.Persist()

	return &raftpb.RequestVoteResponse{
		Term:        int32(s.node.CurrentTerm),
		VoteGranted: granted,
	}, nil

}