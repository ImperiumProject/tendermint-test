package rskip

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

func RoundSkip(sysParams *common.SystemParams, height, round int) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	roundReached := sm.Builder().
		On(common.HeightReached(height), "SkipRounds").
		On(common.RoundReached(round), "roundReached")

	roundReached.MarkSuccess()
	roundReached.On(
		common.DiffCommits(),
		testlib.FailStateLabel,
	)

	cascade := testlib.NewHandlerCascade()
	cascade.AddHandler(common.TrackRoundAll)
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
			sm.InState("roundReached"),
		).Then(
			testlib.Set("DelayedPrevotes").DeliverAll(),
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

	testCase := testlib.NewTestCase(
		"RoundSkipWithPrevotes",
		30*time.Second,
		sm,
		cascade,
	)
	testCase.SetupFunc(common.Setup(sysParams))
	return testCase
}
