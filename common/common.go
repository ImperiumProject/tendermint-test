package common

import (
	"math/rand"

	"github.com/ImperiumProject/imperium/log"
	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/imperium/types"
	"github.com/ImperiumProject/tendermint-test/util"
)

var (
	DefaultOptions = []SetupOption{partition}

	curRoundKey      = "_curRound"
	commitBlockIDKey = "commitBlockId"
	randomReplicaKey = "_randomReplica"
)

type SetupOption func(*testlib.Context)

func Setup(sysParams *SystemParams, options ...SetupOption) func(*testlib.Context) error {
	return func(c *testlib.Context) error {
		c.Vars.Set("n", sysParams.N)
		c.Vars.Set("faults", sysParams.F)
		if len(options) == 0 {
			options = append(options, DefaultOptions...)
		}
		for _, o := range options {
			o(c)
		}
		return nil
	}
}

func PickRandomReplica() SetupOption {
	return func(c *testlib.Context) {
		rI := rand.Intn(c.Replicas.Cap())
		var replica types.ReplicaID
		for i, r := range c.Replicas.Iter() {
			replica = r.ID
			if i == rI {
				break
			}
		}
		c.Vars.Set(randomReplicaKey, replica)
	}
}

func GetRandomReplica(_ *types.Event, c *testlib.Context) (types.ReplicaID, bool) {
	rS, ok := c.Vars.GetString(randomReplicaKey)
	return types.ReplicaID(rS), ok
}

func partition(c *testlib.Context) {
	f := int((c.Replicas.Cap() - 1) / 3)
	partitioner := util.NewGenericPartitioner(c.Replicas)
	partition, _ := partitioner.CreatePartition(
		[]int{1, f, 2 * f},
		[]string{"h", "faulty", "rest"},
	)
	c.Vars.Set("partition", partition)
	c.Logger().With(log.LogParams{
		"partition": partition.String(),
	}).Info("Partitioned replicas")
}

func GetCurRound(ctx *testlib.Context) (int, bool) {
	return ctx.Vars.GetInt(curRoundKey)
}

func GetCommitBlockID(ctx *testlib.Context) (string, bool) {
	return ctx.Vars.GetString(commitBlockIDKey)
}
