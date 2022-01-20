package mainpath

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

func QuorumPrevotes(sysParams *common.SystemParams) *testlib.TestCase {

	cascade := testlib.NewHandlerCascade()

	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("proposal"),
			testlib.DeliverMessage(),
		),
	)

	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageReceive().
				And(testlib.IsMessageToF(common.GetRandomReplica)).
				And(common.IsMessageType(util.Prevote)).
				And(common.IsVoteForProposal("proposal")),
		).Then(
			testlib.Count("prevotesDelivered").Incr(),
		),
	)

	sm := testlib.NewStateMachine()
	init := sm.Builder()

	quorumDelivered := init.On(
		testlib.Count("prevotesDelivered").Geq(2*sysParams.F+1),
		"quorumDelivered",
	)
	quorumDelivered.On(
		testlib.IsMessageSend().
			And(testlib.IsMessageFromF(common.GetRandomReplica)).
			And(common.IsMessageType(util.Precommit)).
			And(common.IsVoteForProposal("proposal")),
		testlib.SuccessStateLabel,
	)

	testcase := testlib.NewTestCase(
		"QuorumPrevotes",
		1*time.Minute,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sysParams, common.PickRandomReplica()))
	return testcase
}

func NilPrevotes(sysParams *common.SystemParams) *testlib.TestCase {
	cascade := testlib.NewHandlerCascade()
	// We don't deliver any proposal and hence we should see that replicas other than the proposer prevote nil.
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageType(util.Proposal)),
		).Then(
			testlib.DropMessage(),
		),
	)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageReceive().
				And(testlib.IsMessageToF(common.GetRandomReplica)).
				And(common.IsMessageType(util.Prevote)).
				And(common.IsNilVote()),
		).Then(
			testlib.Count("nilPrevotesDelivered").Incr(),
		),
	)

	sm := testlib.NewStateMachine()
	init := sm.Builder()

	nilQuorumDelivered := init.On(
		testlib.Count("nilPrevotesDelivered").Geq(2*sysParams.F+1),
		"nilQuorumDelivered",
	)
	nilQuorumDelivered.On(
		testlib.IsMessageSend().
			And(testlib.IsMessageFromF(common.GetRandomReplica)).
			And(common.IsMessageType(util.Precommit)).
			And(common.IsNilVote()),
		testlib.SuccessStateLabel,
	)

	testcase := testlib.NewTestCase(
		"NilPrevotes",
		1*time.Minute,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sysParams, common.PickRandomReplica()))
	return testcase
}
