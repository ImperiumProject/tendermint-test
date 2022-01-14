package common

import (
	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/util"
)

var (
	DefaultOptions = []SetupOption{partition}

	curRoundKey      = "_curRound"
	commitBlockIDKey = "commitBlockId"
)

type SetupOption func(*testlib.Context)

func Setup(sysParams *SystemParams, options ...SetupOption) func(*testlib.Context) error {
	return func(c *testlib.Context) error {
		c.Vars.Set("n", sysParams.N)
		c.Vars.Set("faults", sysParams.F)
		opts := append(DefaultOptions, options...)
		for _, o := range opts {
			o(c)
		}
		return nil
	}
}

func partition(c *testlib.Context) {
	f := int((c.Replicas.Cap() - 1) / 3)
	partitioner := util.NewGenericPartitioner(c.Replicas)
	partition, _ := partitioner.CreatePartition(
		[]int{1, f, 2 * f},
		[]string{"h", "faulty", "rest"},
	)
	c.Vars.Set("partition", partition)
}

func GetCurRound(ctx *testlib.Context) (int, bool) {
	return ctx.Vars.GetInt(curRoundKey)
}

func GetCommitBlockID(ctx *testlib.Context) (string, bool) {
	return ctx.Vars.GetString(commitBlockIDKey)
}
