package raft

import (
	"context"
	"fmt"
	"time"

	"github.com/etcd/raft/raftpb"
)

func (n *Node) broadcastVoteRequestRPC() {

	// compute last log entry
	lastLogIndex := len(n.Log) - 1
	LastLogTerm := n.Log[lastLogIndex].Term

	for id, client := range n.Clients {
		go func(id int, c raftpb.RaftServiceClient) {

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			resp, err := c.RequestVote(ctx, &raftpb.RequestVoteRequest{
				Term: int32(n.CurrentTerm),
				CandidateId: int32(n.Id),
				LastLogIndex: int32(lastLogIndex),
				LastLogTerm: int32(LastLogTerm),
			})
			if err != nil {
				fmt.Printf("Node %d -> RequestVote to %d FAILED: %v\n", n.Id, id, err)
				return
			}
			fmt.Printf("Node %d -> RequestVote to %d SUCCEEDED: %v\n", n.Id, id, resp)
			
			if n.State == Candidate {
				fmt.Printf("Node %d received vote response from %d as %v (term %d)\n", n.Id, id, resp.VoteGranted, resp.Term)
				n.Inbox <- VoteResponse{
					From: id,
					Term: int(resp.Term),
					Granted: resp.VoteGranted,
				}
			}

		}(id, client)
	}

}