package testcases

import (
	"time"

	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/imperium/types"
	"github.com/ImperiumProject/tendermint-test/util"
)

func handler(e *types.Event, c *testlib.Context) ([]*types.Message, bool) {
	if !e.IsMessageSend() {
		return []*types.Message{}, false
	}
	messageID, _ := e.MessageID()
	message, ok := c.MessagePool.Get(messageID)
	if ok {
		return []*types.Message{message}, true
	}
	return []*types.Message{}, true
}

func cond(e *types.Event, c *testlib.Context) bool {
	if !e.IsMessageSend() {
		return false
	}

	message, ok := util.GetMessageFromEvent(e, c)
	if !ok {
		return false
	}
	return message.Type == util.Precommit
}

func DummyTestCaseStateMachine() *testlib.TestCase {
	sm := testlib.NewStateMachine()
	sm.Builder().On(cond, testlib.SuccessStateLabel)

	h := testlib.NewHandlerCascade()
	h.AddHandler(handler)

	testcase := testlib.NewTestCase("DummySM", 30*time.Second, sm, h)
	return testcase
}
