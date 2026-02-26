package raft

import (
	"encoding/json"
	"os"
)

func (s *Storage) Save(state PersistentState) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Create(s.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(state)

}

func (s *Storage) Load() (*PersistentState, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var state PersistentState
	decoder := json.NewDecoder(file)

	err = decoder.Decode(&state)
	if err != nil {
		return nil, err
	}

	return &state, nil

}

func (n *Node) Persist() {
	if n.Storage == nil {
		return
	}
	state := PersistentState{
		Term: n.CurrentTerm,
		VotedFor: n.VotedFor,
		Log: n.Log,
		CommitIndex: n.CommitIndex,
	}
	n.Storage.Save(state)
}

func (s *Storage) SaveSnapshot(snapshot Snapshot) error {

	file, err := os.Create(s.FilePath + ".snapshot")
	if err != nil {
		return err
	}
	defer  file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(snapshot)

}