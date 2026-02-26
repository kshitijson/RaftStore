package raft

import (
	"sync"
	"time"

	"github.com/etcd/raft/raftpb"
	"google.golang.org/grpc"
)

type State int

/*
- iota is a Go counter
- Starts at 0
- Increments by 1 for each line
*/
const (
	Follower State = iota
	Candidate
	Leader
)

type RequestVote struct {
	From int
	Term int
	LastLogIndex int
	LastLogTerm int
}

type VoteResponse struct {
	From    int
	Term    int
	Granted bool
}

type Heartbeat struct {
	From int
	Term int
}

type Node struct {
	Mu sync.Mutex

	Id          int
	State       State
	CurrentTerm int
	VotedFor    int

	Inbox   chan interface{}
	Clients map[int]raftpb.RaftServiceClient
	PeerAddrMap map[int]string
	Server  *grpc.Server
	
	ElectionTimer *time.Timer
	StateCh       chan struct{}
	
	Log []LogEntry
	CommitIndex int
	LastApplied int
	
	NextIndex map[int]int
	MatchIndex map[int]int
	
	KV *KVStore
	
	LeaderId int
	LeaderAddr string
	SelfAddr string

	Storage *Storage
}

type PersistentState struct {
	Term     int
	VotedFor int
	Log []LogEntry
	CommitIndex int
}

type Storage struct {
	mu sync.Mutex
	FilePath string
}

type KVStore struct {
	Mu sync.Mutex

	Data map[string]string
}

type Snapshot struct {
	LastIncludedIndex int
	LastIncludedTerm int
	KVData map[string]string
}

type LogEntry struct {
	Term    int
	Index   int
	Command []byte
}