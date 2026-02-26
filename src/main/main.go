package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/etcd/raft/raftpb"
	"github.com/etcd/src/raft"
	"github.com/etcd/src/rpc"
	"math/rand"
	"time"
)

type Peer struct {
	ID int
	Addr string
}

func init() {
    fmt.Println("INIT of main package reached")
}

func getEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing env " + key)
	}
	return v
}

func getPeers() []Peer {
	
	raw := strings.Split(getEnv("PEERS"), ",")
	peers := make([]Peer, 0)

	for _, p := range raw {
		parts := strings.Split(p, "@")
		id, _ := strconv.Atoi(parts[0])
		peers = append(peers, Peer{
			ID: id,
			Addr: parts[1],
		})
	}

	return peers

}

func main() {

	fmt.Println("Node has Started: MSG for MAIN() ------")

	id, _ := strconv.Atoi(getEnv("NODE_ID"))
	port := getEnv("NODE_PORT")
	peers := getPeers()

	fmt.Println("ID: ", id)
	fmt.Println("Port: ", port)
	fmt.Println("Peers: ", peers)

	node := &raft.Node{
		Id:          id,
		State:       raft.Follower,
		CurrentTerm: 0,
		VotedFor:    -1,
		Inbox:       make(chan interface{}, 100),
		Clients:     make(map[int]raftpb.RaftServiceClient),
		PeerAddrMap: make(map[int]string),
		StateCh:     make(chan struct{}, 1),
		SelfAddr:    getEnv("SELF_ADDR"),
	}

	rand.Seed(time.Now().UnixNano() + int64(node.Id)*1000)

	
	// Initialize KV Store
	node.KV = &raft.KVStore{
		Data: make(map[string]string),
	}
	// Dummy Entry
	node.Log = []raft.LogEntry{
		{Term: 0, Index: 0, Command: []byte("init")},
	}
	
	
	// Persistennt Storage
	os.MkdirAll("/app/data/", 0755)
	node.Storage = &raft.Storage{
		FilePath: "/app/data/wal.json",
	}
	
	state, err := node.Storage.Load()
	if err == nil {
		node.CurrentTerm = state.Term
		node.VotedFor = state.VotedFor
		node.Log = state.Log
		node.CommitIndex = state.CommitIndex
	}
	
	for _, peer := range peers {
		node.PeerAddrMap[peer.ID] = peer.Addr
		node.Clients[peer.ID] = rpc.NewRaftClient(peer.Addr)
	}
	
	rpc.StartGRPCServer(node, ":" + port)

	time.Sleep(2 * time.Second) // allow all nodes to boot
	
	go node.Run()
	
	
	select {}
}


