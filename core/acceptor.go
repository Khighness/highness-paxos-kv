package core

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// @Author KHighness
// @Update 2022-10-15

func (a *BallotNum) GE(b *BallotNum) bool {
	if a.N > b.N {
		return true
	}
	if a.N < b.N {
		return false
	}
	return a.ProposerId >= b.ProposerId
}

// Version defines one modification of a key-value record.
type Version struct {
	mu       sync.Mutex
	acceptor Acceptor
}

// Versions stores all version of a record.
type Versions map[int64]*Version

// KVServer implements the paxos Acceptor API, handling Prepare and Accept request.
type KVServer struct {
	mu      sync.Mutex
	Storage map[string]Versions
}

// Prepare handles Prepare request.
func (s *KVServer) Prepare(c context.Context, r *Proposer) (*Acceptor, error) {
	zap.S().Infof("Acceptor: receive Prepare request: %v", r)

	v := s.getVersionLocked(r.Id)
	defer v.mu.Unlock()

	reply := v.acceptor

	if r.Bal.GE(v.acceptor.LastBal) {
		v.acceptor.LastBal = r.Bal
	}

	return &reply, nil
}

// Accept handles Accept request.
func (s *KVServer) Accept(c context.Context, r *Proposer) (*Acceptor, error) {
	zap.S().Infof("Acceptor: receive Accept request: %v", r)

	v := s.getVersionLocked(r.Id)
	defer v.mu.Unlock()

	d := *v.acceptor.LastBal
	reply := Acceptor{LastBal: &d}

	if r.Bal.GE(v.acceptor.LastBal) {
		v.acceptor.LastBal = r.Bal
		v.acceptor.Val = r.Val
		v.acceptor.VBal = r.Bal
	}

	return &reply, nil
}

func (s *KVServer) getVersionLocked(id *PaxosInstanceId) *Version {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := id.Key
	ver := id.Ver
	versions, ok := s.Storage[key]
	if !ok {
		versions = Versions{}
		s.Storage[key] = versions
	}

	v, ok := versions[ver]
	if !ok {
		versions[ver] = &Version{
			acceptor: Acceptor{
				LastBal: &BallotNum{},
				VBal:    &BallotNum{},
			},
		}
		v = versions[ver]
	}

	v.mu.Lock()
	return v
}
