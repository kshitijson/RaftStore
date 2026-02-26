package raft

import (
	"errors"
	"time"
)

func (n *Node) Propose(command []byte) error {

	n.Mu.Lock()

	if n.State != Leader {
		n.Mu.Unlock()
		return errors.New("not leader")
	}

	index := len(n.Log)
	term := n.CurrentTerm

	entry := LogEntry{
		Term:    term,
		Index:   index,
		Command: command,
	}

	n.Log = append(n.Log, entry)
	n.Mu.Unlock()

	n.Persist()

	// Wait for the entry to be committed by a majority
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			return errors.New("timeout: entry not committed (quorum unavailable)")
		default:
			n.Mu.Lock()
			committed := n.CommitIndex >= index
			stillLeader := n.State == Leader && n.CurrentTerm == term
			n.Mu.Unlock()

			if committed {
				return nil
			}
			if !stillLeader {
				return errors.New("lost leadership before entry was committed")
			}
			time.Sleep(10 * time.Millisecond)
		}
	}

}