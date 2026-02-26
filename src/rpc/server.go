package rpc

import (
	"log"
	"net"

	"github.com/etcd/raft/raftpb"
	"github.com/etcd/src/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartGRPCServer(n *raft.Node, port string) {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to Listen: %v", err)
	}

	server := grpc.NewServer()
	raftpb.RegisterRaftServiceServer(server, &RaftServer{node: n})

	raftpb.RegisterKVServiceServer(server, &RaftServer{node: n})

	log.Printf("Node %d listening on %s\n", n.Id, port)

	go server.Serve(lis)

}

func NewRaftClient(addr string) raftpb.RaftServiceClient {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	return raftpb.NewRaftServiceClient(conn)
}