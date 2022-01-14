package common

import (
	"strconv"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/imperium/types"
	"github.com/ImperiumProject/tendermint-test/util"
)

func IsCommit() testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if eType.T == "Committing block" {
			blockID, ok := eType.Params["block_id"]
			if ok {
				c.Vars.Set(commitBlockIDKey, blockID)
			}
		}
		return ok && eType.T == "Committing block"
	}
}

func IsMessageFromRound(round int) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		return m.Round() == round
	}
}

func IsVoteFromPart(partS string) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		if m.Type != util.Precommit && m.Type != util.Prevote {
			return false
		}

		partition, ok := getPartition(c)
		if !ok {
			return false
		}
		part, ok := partition.GetPart(partS)
		if !ok {
			return false
		}
		val, ok := util.GetVoteValidator(m)
		if !ok {
			return false
		}
		return part.ContainsVal(val)
	}
}

func IsVoteFromFaulty() testlib.Condition {
	return IsVoteFromPart("faulty")
}

func getPartition(c *testlib.Context) (*util.Partition, bool) {
	p, exists := c.Vars.Get("partition")
	if !exists {
		return nil, false
	}
	partition, ok := p.(*util.Partition)
	return partition, ok
}

func IsMessageFromPart(partS string) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		partition, ok := getPartition(c)
		if !ok {
			return false
		}
		part, ok := partition.GetPart(partS)
		if !ok {
			return false
		}
		return part.Contains(m.From)
	}
}

func IsMessageToPart(partS string) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		partition, ok := getPartition(c)
		if !ok {
			return false
		}
		part, ok := partition.GetPart(partS)
		if !ok {
			return false
		}
		return part.Contains(m.To)
	}
}

func IsMessageType(t util.MessageType) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		tMessage, ok := util.GetParsedMessage(message)
		if !ok {
			return false
		}
		return tMessage.Type == t
	}
}

// RoundReached returns true if all replicas have reached the specified round
// Should be used with TrackRound handler!
func RoundReached(r int) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		curRound, ok := c.Vars.GetInt(curRoundKey)
		if ok && curRound >= r {
			return true
		}
		return false
	}
}

func TwoFMinus1() func(*types.Event, *testlib.Context) (int, bool) {
	return func(e *types.Event, c *testlib.Context) (int, bool) {
		f, ok := c.Vars.GetInt("faults")
		if !ok {
			return 0, false
		}
		return 2*f + 1, true
	}
}

func IsVoteForProposal(proposalLabel string) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		proposal, ok := c.Vars.GetString(proposalLabel)
		if !ok {
			return false
		}
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		tMsg, ok := util.GetParsedMessage(message)
		if !ok {
			return false
		}
		voteBlockID, ok := util.GetVoteBlockIDS(tMsg)
		if !ok {
			return false
		}
		return voteBlockID == proposal
	}
}

func IsProposalEq(proposalLabel string) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		proposal, ok := c.Vars.GetString(proposalLabel)
		if !ok {
			return false
		}
		message, ok := c.GetMessage(e)
		if !ok {
			return false
		}
		tMsg, ok := util.GetParsedMessage(message)
		if !ok {
			return false
		}
		proposalBlockID, ok := util.GetProposalBlockIDS(tMsg)
		if !ok {
			return false
		}
		return proposalBlockID == proposal
	}
}

func IsFromHeight(height int) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		m, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		return m.Height() == height
	}
}

func HeightReached(h int) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if !ok {
			return false
		}
		if eType.T != "newStep" {
			return false
		}
		heightS, ok := eType.Params["height"]
		if !ok {
			return false
		}
		height, err := strconv.Atoi(heightS)
		if err != nil {
			return false
		}
		return height == h
	}
}

func IsEventNewRound(r int) testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		eType, ok := e.Type.(*types.GenericEventType)
		if !ok {
			return false
		}
		if eType.T != "newStep" {
			return false
		}
		roundS, ok := eType.Params["round"]
		if !ok {
			return false
		}
		round, err := strconv.Atoi(roundS)
		if err != nil {
			return false
		}
		return round >= r
	}
}