package rpc

import (
	"github.com/etcd/raft/raftpb"
	"github.com/etcd/src/raft"
)

type RaftServer struct {
	raftpb.UnimplementedRaftServiceServer
	raftpb.UnimplementedKVServiceServer
	node *raft.Node
}