package core

import (
	"context"
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/khighness/highness-paxos-kv/pkg/logging"
)

// @Author KHighness
// @Update 2022-10-15

// GE compares the ballot number with another BallotNum.
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

// ServeAcceptors starts a gRPC server for every acceptor.
func ServeAcceptors(acceptorIds []int64) []*grpc.Server {
	var servers []*grpc.Server

	for _, aid := range acceptorIds {
		addr := fmt.Sprintf(":%d", AcceptorBasePort+int(aid))

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			zap.S().Fatalf("listen: %s %v", addr, err)
		}

		server := grpc.NewServer()
		RegisterPaxosKVServer(server, &KVServer{Storage: map[string]Versions{}})
		reflection.Register(server)
		zap.S().Infof("Acceptor-%d is serving on %s", aid, addr)
		servers = append(servers, server)
		go server.Serve(listener)
	}

	return servers
}
