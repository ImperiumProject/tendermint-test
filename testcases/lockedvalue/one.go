package lockedvalue

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

func ExpectUnlock(sysParams *common.SystemParams) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	init := sm.Builder()

	roundOne := init.On(common.RoundReached(1), "RoundOne")

	roundOne.On(
		testlib.IsMessageSend().
			And(common.IsMessageFromRound(1).Not()).
			And(common.IsMessageFromRound(0).Not()).
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("zeroProposal")),
		testlib.FailStateLabel,
	)
	roundOne.On(
		testlib.IsMessageSend().
			And(common.IsMessageFromRound(1).Not()).
			And(common.IsMessageFromRound(0).Not()).
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("zeroProposal").Not()),
		testlib.SuccessStateLabel,
	)
	init.On(
		common.IsCommit(), testlib.FailStateLabel,
	)

	cascade := testlib.NewHandlerCascade()
	cascade.AddHandler(common.TrackRoundAll)
	// Change faulty replicas votes to nil
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsVoteFromFaulty()),
		).Then(common.ChangeVoteToNil()),
	)
	// Record round 0 proposal
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("zeroProposal"),
		),
	)
	// Do not deliver votes from "h"
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsVoteFromPart("h")),
		).Then(
			testlib.Set("zeroDelayedPrevotes").Store(),
			testlib.DropMessage(),
		),
	)
	// For higher rounds, we do not deliver proposal until we see a new one
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0).Not()).
				And(common.IsProposalEq("zeroProposal")),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"ExpectUnlock",
		1*time.Minute,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sysParams))
	return testcase
}
