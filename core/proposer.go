package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @Author KHighness
// @Update 2022-10-15

var (
	ErrNoEnoughQuorum = errors.New("no enough quorum")
	AcceptorBasePort  = 3333
)

// RunPaxos executes the paxos phase-1 and phase-2 to establish a value.
// It returns the established value, which may be a voted value that is not `val`.
//
// If `val` is not nil, it acts as a writing operation.
// If `val` is nil, it acts as a reading operation.
func (p *Proposer) RunPaxos(acceptorIds []int64, val *Value) *Value {
	quorum := len(acceptorIds)/2 + 1

	for {
		p.Val = nil

		maxVotedVal, higherBal, err := p.Phase1(acceptorIds, quorum)
		if err != nil {
			zap.S().Error("Proposer: failed to run phase-1, highest ballot: %v, increment ballot and retry", higherBal)
			p.Bal.N = higherBal.N + 1
			continue
		}

		if maxVotedVal == nil {
			zap.S().Infof("Proposer: no voted value seen, propose my value: %v", val)
		} else {
			val = maxVotedVal
		}

		if val == nil {
			zap.S().Infof("Proposer: no value to propose in phase-2, quit")
			return nil
		}

		p.Val = val
		zap.S().Infof("Proposer: proposer chose value to propose: %s", p.Val)

		higherBal, err = p.Phase2(acceptorIds, quorum)
		if err != nil {
			zap.S().Infof("Proposer: failed to run phase-2, highest ballot: %v, increment ballot and retry", higherBal)
			p.Bal.N = higherBal.N + 1
			continue
		}

		zap.S().Infof("Proposer: value is voted by a quorum and has been safe: %v", p.Val)
		return p.Val
	}
}

// Phase1 runs paxos phase-1 on the specified acceptorIds.
// If a higher ballot number is seen and phase-1 failed to constitute a quorum,
// the highest ballot and a ErrNoEnoughQuorum will be returned.
func (p *Proposer) Phase1(acceptorIds []int64, quorum int) (*Value, *BallotNum, error) {
	replies := p.rpcToAll(acceptorIds, "Prepare")

	var count int
	higherBal := *p.Bal
	maxVoted := &Acceptor{VBal: &BallotNum{}}

	for _, r := range replies {
		zap.S().Infof("Proposer: handling Prepare Reply: %v", r)

		if !p.Bal.GE(r.LastBal) {
			if r.LastBal.GE(&higherBal) {
				higherBal = *r.LastBal
			}
			continue
		}

		if r.VBal.GE(maxVoted.VBal) {
			maxVoted = r
		}

		count += 1
		if count == quorum {
			return maxVoted.Val, nil, nil
		}
	}

	return nil, &higherBal, ErrNoEnoughQuorum
}

// Phase2 runs paxos phase-2 on the specified acceptorIds.
// If a higher ballot number is seen and phase-1 failed to constitute a quorum,
// the highest ballot and a ErrNoEnoughQuorum will be returned.
func (p *Proposer) Phase2(acceptorIds []int64, quorum int) (*BallotNum, error) {
	replies := p.rpcToAll(acceptorIds, "Accept")

	var count int
	higherBal := *p.Bal
	for _, r := range replies {
		zap.S().Infof("Proposer: handling Accept reply: %v", r)

		if !p.Bal.GE(r.LastBal) {
			if r.LastBal.GE(&higherBal) {
				higherBal = *r.LastBal
			}
			continue
		}

		count += 1
		if count == quorum {
			return nil, nil
		}
	}

	return &higherBal, ErrNoEnoughQuorum
}

// rpcToAll sends Prepare or Accept RPC to the specified Acceptors.
func (p *Proposer) rpcToAll(acceptorIds []int64, action string) []*Acceptor {
	var replies []*Acceptor

	for _, aid := range acceptorIds {
		address := fmt.Sprintf("127.0.0.1:%d", AcceptorBasePort+int(aid))
		conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			zap.S().Panicf("failed to connect: %v", err)
		}

		defer conn.Close()
		c := NewPaxosKVClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		var reply *Acceptor
		switch action {
		case "Prepare":
			reply, err = c.Prepare(ctx, p)
		case "Accept":
			reply, err = c.Accept(ctx, p)
		}
		if err != nil {
			zap.S().Errorf("Proposer: %s failure from Acceptor-%d: %v", action, aid, err)
		}
		zap.S().Infof("Proposer: receive %s reply from Acceptor-%d: %v", action, aid, reply)

		if reply != nil {
			replies = append(replies, reply)
		}
	}

	return replies
}
