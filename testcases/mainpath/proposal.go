package mainpath

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

func ProposePrevote(sp *common.SystemParams) *testlib.TestCase {
	sm := testlib.NewStateMachine()

	init := sm.Builder()
	proposalDelivered := init.On(
		testlib.IsMessageReceive().
			And(common.IsMessageType(util.Proposal)).
			And(testlib.IsMessageToF(common.GetRandomReplica)),
		"ProposalDelivered",
	)
	proposalDelivered.On(
		testlib.IsMessageSend().
			And(common.IsMessageType(util.Prevote)).
			And(testlib.IsMessageFromF(common.GetRandomReplica)),
		testlib.SuccessStateLabel,
	)

	cascade := testlib.NewHandlerCascade()
	testcase := testlib.NewTestCase(
		"ProposePrevote",
		15*time.Second,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sp, common.PickRandomReplica()))
	return testcase
}

func ProposalNilPrevote(sp *common.SystemParams) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	init := sm.Builder()
	init.On(
		testlib.IsMessageSend().
			And(testlib.IsMessageFromF(common.GetRandomReplica)).
			And(common.IsMessageType(util.Prevote)).
			And(common.IsNotNilVote()),
		testlib.FailStateLabel,
	)

	cascade := testlib.NewHandlerCascade()

	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(testlib.IsMessageToF(common.GetRandomReplica)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"ProposalNilPrevote",
		30*time.Second,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sp, common.PickRandomReplica()))
	return testcase
}
