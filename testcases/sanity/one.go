package sanity

import (
	"fmt"
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/imperium/types"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

func quorumCond() testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		message, ok := util.GetMessageFromEvent(e, c)
		if !ok {
			return false
		}
		partitionI, _ := c.Vars.Get("partition")
		partition := partitionI.(*util.Partition)
		faulty, _ := partition.GetPart("faulty")
		if faulty.Contains(message.From) {
			return false
		}

		if message.Type != util.Prevote {
			return false
		}
		round := message.Round()
		blockID, _ := util.GetVoteBlockIDS(message)

		replicaVotekey := fmt.Sprintf("votesReplica_%d_%s", round, message.From)
		blockVoteKey := fmt.Sprintf("votesBlock_%d_%s", round, blockID)

		if !c.Vars.Exists(replicaVotekey) {
			c.Vars.Set(replicaVotekey, blockID)
		}
		counter, ok := c.Vars.GetCounter(blockVoteKey)
		if !ok {
			c.Vars.SetCounter(blockVoteKey)
			counter, _ = c.Vars.GetCounter(blockVoteKey)
		}
		counter.Incr()
		if round != 0 {
			return findIntersection(c)
		}
		return false
	}
}

func findIntersection(ctx *testlib.Context) bool {
	f, _ := ctx.Vars.GetInt("faults")
	votes := make(map[string]map[string]bool)
	for _, r := range ctx.Replicas.Iter() {
		if blockID, ok := ctx.Vars.GetString(fmt.Sprintf("votesReplica_1_%s", r.ID)); ok {
			_, ok := votes[blockID]
			if !ok {
				votes[blockID] = make(map[string]bool)
			}
			votes[blockID][string(r.ID)] = true
		}
	}
	var quorumBlock string = ""
	for blockID := range votes {
		if count, ok := ctx.Vars.GetCounter(fmt.Sprintf("votesBlock_1_%s", blockID)); ok {
			if count.Value() >= 2*f+1 {
				quorumBlock = blockID
				break
			}
		}
	}
	if quorumBlock == "" {
		return false
	}

	if !ctx.Vars.Exists(fmt.Sprintf("votesBlock_1_%s", quorumBlock)) {
		return false
	}
	for replica := range votes[quorumBlock] {
		if !ctx.Vars.Exists(fmt.Sprintf("votesReplica_0_%s", replica)) {
			return false
		}
	}
	return true
}

// States:
// 	1. Skip rounds by not delivering enough precommits to the replicas
// 		1.1. Ensure one faulty replica prevotes and precommits nil
// 	2. Check that in the new round there is a quorum intersection of f+1
// 		2.1 Record the votes on the proposal to check for quorum intersection (Proposal should be same in both rounds)
func OneTestCase(sysParams *common.SystemParams) *testlib.TestCase {

	sm := testlib.NewStateMachine()
	sm.Builder().On(quorumCond(), testlib.SuccessStateLabel)

	handler := testlib.NewHandlerCascade()
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().And(common.IsVoteFromFaulty()),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Precommit)).
				And(common.IsMessageToPart("faulty")).
				And(testlib.CountTo("voteCount").Lt(2*sysParams.F-1)),
		).Then(
			testlib.CountTo("voteCount").Incr(),
			testlib.DeliverMessage(),
		),
	)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Precommit)).
				And(common.IsMessageToPart("faulty")).
				And(testlib.CountTo("voteCount").Geq(2*sysParams.F - 1)),
		).Then(
			testlib.DropMessage(),
		),
	)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Precommit)).
				And(testlib.CountTo("voteCount").Leq(2*sysParams.F-1)),
		).Then(
			testlib.CountTo("voteCount").Incr(),
			testlib.DeliverMessage(),
		),
	)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Precommit)).
				And(testlib.CountTo("voteCount").Gt(2*sysParams.F - 1)),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase("QuorumIntersection", 50*time.Second, sm, handler)
	testcase.SetupFunc(common.Setup(sysParams))

	return testcase
}
