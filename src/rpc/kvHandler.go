package rpc

import (
	"context"
	"fmt"

	"github.com/etcd/raft/raftpb"
	"github.com/etcd/src/raft"
)

func (s *RaftServer) Put(
	ctx context.Context,
	req *raftpb.PutRequest,
) (*raftpb.PutResponse, error) {

	n := s.node

	fmt.Printf("Leader address: %s\n", n.LeaderAddr)

	if n.State != raft.Leader {
		return &raftpb.PutResponse{
			Success: false,
			LeaderAddr: n.LeaderAddr,
		}, nil
	}

	cmd := fmt.Sprintf("PUT %s %s", req.Key, req.Value)

	err := n.Propose([]byte(cmd))
	if err != nil {
		return &raftpb.PutResponse{
			Success: false,
			LeaderAddr: n.LeaderAddr,
			}, nil
	}

	return &raftpb.PutResponse{Success: true}, nil
}

func (s *RaftServer) Get(
	ctx context.Context,
	req *raftpb.GetRequest,
) (*raftpb.GetResponse, error) {

	n := s.node

	n.KV.Mu.Lock()
	value := n.KV.Data[req.Key]
	n.KV.Mu.Unlock()

	return &raftpb.GetResponse{
		Value: value,
		LeaderAddr: n.LeaderAddr,
	}, nil

}