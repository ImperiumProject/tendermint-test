package invariant

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
)

func NotNilDecide(sp *common.SystemParams) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	init := sm.Builder()
	init.MarkSuccess()
	init.On(
		common.IsNilCommit(),
		testlib.FailStateLabel,
	)

	cascade := testlib.NewHandlerCascade()
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
				And(common.IsVoteFromPart("h")),
		).Then(
			testlib.DropMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"NotNilDecide",
		2*time.Minute,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sp))
	return testcase
}
