package rskip

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

func RoundSkip(sysParams *common.SystemParams, height, round int) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	sm.Builder().
		On(common.HeightReached(height), "SkipRounds").
		On(common.RoundReached(round), "DeliverDelayed").
		On(testlib.Set("DelayedPrevotes").Count().Eq(0), testlib.SuccessStateLabel)

	cascade := testlib.NewHandlerCascade()
	cascade.AddHandler(common.TrackRound)
	cascade.AddHandler(
		testlib.If(
			common.IsFromHeight(height).Not(),
		).Then(
			testlib.DeliverMessage(),
		),
	)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsFromHeight(height)).
				And(common.IsVoteFromFaulty()),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsFromHeight(height)).
				And(common.IsMessageFromPart("h")).
				And(common.IsMessageType(util.Prevote)),
		).Then(
			testlib.Set("DelayedPrevotes").Store(),
			testlib.DropMessage(),
		),
	)
	cascade.AddHandler(
		testlib.If(
			sm.InState("DeliverDelayed"),
		).Then(
			testlib.Set("DelayedPrevotes").DeliverAll(),
		),
	)

	testCase := testlib.NewTestCase(
		"RoundSkipWithPrevotes",
		30*time.Second,
		sm,
		cascade,
	)
	testCase.SetupFunc(common.Setup(sysParams))
	return testCase
}
