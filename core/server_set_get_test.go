package core

import (
	"testing"
)

// @Author KHighness
// @Update 2022-10-16

func TestAcceptor_SetAndGetByKey(t *testing.T) {
	// start up 3 acceptors
	acceptorIds := []int64{0, 1, 2}
	servers := ServeAcceptors(acceptorIds)
	defer func() {
		for _, server := range servers {
			server.Stop()
		}
	}()

	// set k-0 = 5
	{
		prop := Proposer{
			Id: &PaxosInstanceId{
				Key: "k",
				Ver: 0,
			},
			Bal: &BallotNum{N: 0, ProposerId: 2},
		}
		value := prop.RunPaxos(acceptorIds, &Value{Vi64: 5})
		t.Logf("set: %v", value)
	}

	// get k-0
	{
		prop := Proposer{
			Id: &PaxosInstanceId{
				Key: "k",
				Ver: 0,
			},
			Bal: &BallotNum{N: 0, ProposerId: 2},
		}
		value := prop.RunPaxos(acceptorIds, nil)
		t.Logf("get: %v", value)
	}

	// set k-1 = 6
	{
		prop := Proposer{
			Id: &PaxosInstanceId{
				Key: "k",
				Ver: 1,
			},
			Bal: &BallotNum{N: 0, ProposerId: 2},
		}
		value := prop.RunPaxos(acceptorIds, &Value{Vi64: 6})
		t.Logf("set: %v", value)
	}

	// get k-1
	{
		prop := Proposer{
			Id: &PaxosInstanceId{
				Key: "k",
				Ver: 1,
			},
			Bal: &BallotNum{N: 0, ProposerId: 2},
		}
		value := prop.RunPaxos(acceptorIds, nil)
		t.Logf("get: %v", value)
	}
}
