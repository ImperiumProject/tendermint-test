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
	init.On(
		testlib.IsMessageSend().
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("zeroProposal")),
		testlib.SuccessStateLabel,
	)

	cascade := testlib.NewHandlerCascade()
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("zeroProposal"),
			testlib.DeliverMessage(),
		),
	)
	testcase := testlib.NewTestCase(
		"ProposePrevote",
		15*time.Second,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
