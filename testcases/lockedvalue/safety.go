package lockedvalue

import (
	"time"

	"github.com/ImperiumProject/imperium/log"
	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/imperium/types"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/util"
)

// Setup function for the test case. Replicas are partitioned into
// "h" - 1, "faulty" - f, "delay" - f and "rest" - f
func safetySetup(c *testlib.Context) {
	f := int((c.Replicas.Cap() - 1) / 3)
	partitioner := util.NewGenericPartitioner(c.Replicas)
	partition, _ := partitioner.CreatePartition(
		[]int{1, f, f, f},
		[]string{"h", "faulty", "delay", "rest"},
	)
	c.Vars.Set("partition", partition)
	c.Logger().With(log.LogParams{
		"partition": partition.String(),
	}).Info("Partitioned replicas")
}

func DifferentDecisions(sysParams *common.SystemParams) *testlib.TestCase {
	sm := testlib.NewStateMachine()
	init := sm.Builder()

	init.On(common.IsCommit(), testlib.FailStateLabel)
	precommitOld := init.On(
		testlib.IsMessageSend().
			And(common.IsMessageType(util.Precommit)).
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("zeroProposal")),
		"PrecommittedOld",
	)
	precommitOld.MarkSuccess()

	precommitOld.On(
		testlib.IsMessageSend().
			And(common.IsMessageType(util.Precommit)).
			And(common.IsVoteFromPart("h")).
			And(common.IsVoteForProposal("newProposal")),
		testlib.FailStateLabel,
	)
	precommitOld.On(
		diffCommits(),
		testlib.FailStateLabel,
	)

	cascade := testlib.NewHandlerCascade()
	cascade.AddHandler(common.TrackRound)

	cascade.AddHandler(
		testlib.If(
			common.IsCommit(),
		).Then(
			recordCommit(),
		),
	)

	// Store the correct precommit messages of round 0 from replica "h" or "faulty" to all replicas in "delay"
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Precommit)).
				And(common.IsVoteFromPart("h").Or(common.IsVoteFromFaulty())).
				And(common.IsMessageToPart("delay")),
		).Then(
			testlib.Set("zeroCorrectPrecommit").Store(),
			testlib.DropMessage(),
		),
	)
	// Store the correct prevotes of round 0 from "h" and "rest" to "delay"
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Prevote)).
				And(common.IsVoteFromPart("h").Or(common.IsVoteFromPart("rest"))).
				And(common.IsMessageToPart("delay")),
		).Then(
			testlib.Set("zeroCorrectPrevotes").Store(),
			testlib.DropMessage(),
		),
	)
	// All other votes to "delay" from other than "delay" are dropped
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageType(util.Precommit).Or(common.IsMessageType(util.Prevote))).
				And(common.IsMessageFromPart("delay").Not()).
				And(common.IsMessageToPart("delay")),
		).Then(
			testlib.DropMessage(),
		),
	)
	// All messages from other rounds to "delay" are also dropped
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsConsensusMessage()).
				And(common.IsMessageFromRound(0).Not()).
				And(common.IsMessageToPart("delay")),
		).Then(
			testlib.DropMessage(),
		),
	)
	// All votes from "delay" which are not to "delay" are dropped
	// cascade.AddHandler(
	// 	testlib.If(
	// 		testlib.IsMessageSend().
	// 			And(common.IsVoteFromPart("delay")).
	// 			And(common.IsMessageToPart("delay").Not()),
	// 	).Then(
	// 		testlib.DropMessage(),
	// 	),
	// )

	// Votes from "h" for the "newProposal" are delivered
	// cascade.AddHandler(
	// 	testlib.If(
	// 		testlib.IsMessageSend().
	// 			And(common.IsVoteFromPart("h")).
	// 			And(common.IsVoteForProposal("newProposal")),
	// 	).Then(
	// 		testlib.DeliverMessage(),
	// 	),
	// )

	// Votes from "h" for round 0 are dropped
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsVoteFromPart("h")),
		).Then(
			testlib.DropMessage(),
		),
	)

	// Votes from "faulty" are changed to nil if not seen new proposal, to new proposal otherwise
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsVoteFromFaulty()),
		).Then(
			changeVote(),
		),
	)
	// Record round 0 proposal
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0)).
				And(common.IsMessageType(util.Proposal)),
		).Then(
			common.RecordProposal("zeroProposal"),
			testlib.RecordMessageAs("zeroProposalMessage"),
			testlib.DeliverMessage(),
		),
	)
	// For higher rounds, we do not deliver proposal until we see a new one
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0).Not()).
				And(common.IsMessageType(util.Proposal)).
				And(common.IsProposalEq("zeroProposal")),
		).Then(
			testlib.DropMessage(),
		),
	)
	// Record the new proposal message
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsMessageFromRound(0).Not()).
				And(common.IsMessageType(util.Proposal)).
				And(common.IsProposalEq("zeroProposal").Not()),
		).Then(
			common.RecordProposal("newProposal"),
			testlib.RecordMessageAs("newProposalMessage"),
			testlib.DeliverMessage(),
		),
	)
	// Once "h" precommits new proposal, deliver votes to "delay"
	cascade.AddHandler(
		testlib.If(
			testlib.IsMessageSend().
				And(common.IsVoteFromPart("h")).
				And(common.IsVoteForProposal("newProposal")),
		).Then(
			testlib.Set("zeroCorrectPrevotes").DeliverAll(),
			testlib.Set("zeroCorrectPrecommit").DeliverAll(),
			testlib.DeliverMessage(),
		),
	)

	testcase := testlib.NewTestCase(
		"DifferentDecisions",
		3*time.Minute,
		sm,
		cascade,
	)
	testcase.SetupFunc(common.Setup(sysParams, safetySetup))
	return testcase
}

func recordCommit() testlib.Action {
	return func(e *types.Event, c *testlib.Context) []*types.Message {
		eType, ok := e.Type.(*types.GenericEventType)
		if ok && eType.T == "Committing block" {
			blockID, ok := eType.Params["block_id"]
			if ok {
				if c.Vars.Exists("commitOne") {
					c.Vars.Set("commitTwo", blockID)
				} else {
					c.Vars.Set("commitOne", blockID)
				}
			}
		}
		return []*types.Message{}
	}
}

func diffCommits() testlib.Condition {
	return func(e *types.Event, c *testlib.Context) bool {
		return c.Vars.Exists("commitOne") && c.Vars.Exists("commitTwo")
	}
}
