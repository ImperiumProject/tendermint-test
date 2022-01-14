package sanity

import (
	"time"

	"github.com/ImperiumProject/imperium/log"
	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/imperium/types"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

func threeSetup(c *testlib.Context) {
	faults := int((c.Replicas.Cap() - 1) / 3)
	partitioner := util.NewGenericPartitioner(c.Replicas)
	partition, _ := partitioner.CreatePartition([]int{faults + 1, 2*faults - 1, 1}, []string{"toLock", "toNotLock", "faulty"})
	c.Logger().With(log.LogParams{
		"partition": partition.String(),
	}).Info("Created partition")
	c.Vars.Set("partition", partition)
}

func commitEqOldProposal() testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		commitBlock, ok := common.GetCommitBlockID(c)
		if !ok {
			return false
		}
		oldProposalI, ok := c.Vars.Get("oldProposal")
		if !ok {
			return false
		}
		oldProposalM, ok := oldProposalI.(*types.Message)
		if !ok {
			return false
		}
		tMsg, ok := util.GetParsedMessage(oldProposalM)
		if !ok {
			return false
		}
		oldProposal, ok := util.GetProposalBlockIDS(tMsg)
		if !ok {
			return false
		}
		c.Logger().With(log.LogParams{
			"commit_blockID": commitBlock,
			"proposal":       oldProposal,
		}).Info("Checking assertion")
		return oldProposal == commitBlock
	}
}

func ThreeTestCase(sysParams *common.SystemParams) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	initialState := sm.Builder()
	initialState.On(common.IsCommit(), testlib.FailStateLabel)
	round1 := initialState.On(common.RoundReached(1), "round1")
	round1.On(common.IsCommit(), testlib.FailStateLabel)
	round2 := round1.On(common.RoundReached(2), "round2")
	round2.On(common.IsCommit().And(commitEqOldProposal()), testlib.SuccessStateLabel)

	handler := testlib.NewHandlerCascade()
	handler.AddHandler(common.TrackRound)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsVoteFromFaulty()),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageToPart("toNotLock")).
				And(common.IsMessageType(util.Prevote)).
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
				And(common.IsMessageToPart("toNotLock")).
				And(common.IsMessageType(util.Prevote)).
				And(testlib.CountTo("voteCount").Geq(2*sysParams.F - 1)),
		).Then(
			testlib.DropMessage(),
		),
	)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			testlib.RecordMessageAs("oldProposal"),
		),
	)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(1)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase("LockedValueCheck", 30*time.Second, sm, handler)
	testcase.SetupFunc(common.Setup(sysParams, threeSetup))

	return testcase
}
