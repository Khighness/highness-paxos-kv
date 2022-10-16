package core

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// @Author KHighness
// @Update 2022-10-16

func TestAcceptor_Accept_LastBal(t *testing.T) {
	r := require.New(t)

	kvServer := KVServer{Storage: map[string]Versions{}}
	proposer := &Proposer{
		Id: &PaxosInstanceId{
			Key: "k",
			Ver: 0,
		},
		Bal: &BallotNum{N: 1},
	}

	reply, err := kvServer.Accept(nil, proposer)
	r.Nil(err)

	version := kvServer.Storage["k"][0]

	version.acceptor.LastBal.N = 100
	r.Equal(int64(0), reply.LastBal.N)
}
