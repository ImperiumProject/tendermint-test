package byzantine

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

func ForeverLaggingReplica(sp *common.SystemParams) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	init := sm.Builder()
	init.On(common.IsCommit(), testlib.FailStateLabel)

	allowCatchUp := init.On(common.RoundReached(5), "allowCatchUp")
	allowCatchUp.On(
		common.IsCommit(),
		testlib.SuccessStateLabel,
	)
	allowCatchUp.On(
		common.DiffCommits(),
		testlib.FailStateLabel,
	)

	cascade := testlib.NewHandlerCascade()
	cascade.AddHandler(common.TrackRoundTwoThirds)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsVoteFromFaulty()),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsVoteFromPart("h")),
		).Then(
			testlib.DropMessage(),
		),
	)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Prevote).Or(common.IsMessageType(util.Precommit))).
				And(sm.InState("allowCatchUp").Not()),
		).Then(
			testlib.DropMessage(),
		),
	)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageToPart("h")).
				And(common.IsMessageFromCurRound()),
		).Then(
			testlib.DeliverMessage(),
		),
	)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageToPart("h")).
				And(common.IsMessageType(util.Prevote).Or(common.IsMessageType(util.Precommit))).
				And(common.MessageCurRoundGt(2)),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"LaggingReplica",
		25*time.Minute,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
