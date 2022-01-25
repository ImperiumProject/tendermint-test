package lockedvalue

import (
	"time"

	"github.com/ImperiumProject/imperium/log"
	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/imperium/types"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

type twoFilters struct{}

func (twoFilters) changeProposalToNil(e *types.Event, c *testlib.Context) []*types.Message {
	message, _ := c.GetMessage(e)
	tMsg, ok := util.GetParsedMessage(message)
	if !ok {
		return []*types.Message{}
	}
	replica, _ := c.Replicas.Get(tMsg.From)
	newProp, err := util.ChangeProposalBlockIDToNil(replica, tMsg)
	if err != nil {
		c.Logger().With(log.LogParams{"error": err}).Error("Failed to change proposal")
		return []*types.Message{message}
	}
	newMsgB, err := newProp.Marshal()
	if err != nil {
		c.Logger().With(log.LogParams{"error": err}).Error("Failed to marshal changed proposal")
		return []*types.Message{message}
	}
	return []*types.Message{c.NewMessage(message, newMsgB)}
}

// States:
// 	1. Ensure replicas skip round by not delivering enough precommits
//		1.1 One replica prevotes and precommits nil
// 	2. In the next round change the proposal block value
// 	3. Replicas should prevote and precommit the earlier block and commit
func LockedCommit(sysParams *common.SystemParams) *testlib.TestCase {

	filters := twoFilters{}

	sm := testlib.NewStateMachine()
	initialState := sm.Builder()
	initialState.On(common.IsCommit(), testlib.FailStateLabel)
	round1 := initialState.On(common.RoundReached(1), "round1")
	round1.On(common.IsCommit(), testlib.SuccessStateLabel)
	round1.On(common.RoundReached(2), testlib.FailStateLabel)

	handler := testlib.NewHandlerCascade()
	handler.AddHandler(common.TrackRoundAll)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().And(common.IsVoteFromFaulty()),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	// Blanket change of all precommits in round 0 to nil,
	// We expect replicas to lock onto the proposal and this is just to ensure they move to the next round
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Precommit)),
		).Then(
			common.ChangeVoteToNil(),
		),
	)
	handler.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(1)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			filters.changeProposalToNil,
		),
	)

	testcase := testlib.NewTestCase("WrongProposal", 30*time.Second, sm, handler)
	testcase.SetupFunc(common.Setup(sysParams))

	return testcase
}
