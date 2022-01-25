package byzantine

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
)

func CrashReplica(sp *common.SystemParams) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	init := sm.Builder()
	roundOne := init.On(
		common.RoundReached(1),
		"roundOne",
	)
	roundOne.On(
		common.IsCommit(),
		testlib.SuccessStateLabel,
	)

	cascade := testlib.NewHandlerCascade()
	cascade.AddHandler(common.TrackRoundTwoThirds)
	cascade.AddHandler(
		testlib.If(
			testlib.Once(sm.InState("roundOne")),
		).Then(
			testlib.StopReplica(common.RandomReplicaFromPart("faulty")),
		),
	)
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
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

	testcase := testlib.NewTestCase(
		"CrashReplica",
		2*time.Minute,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
