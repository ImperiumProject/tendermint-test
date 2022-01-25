package common

import (
	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/imperium/types"
)

func RandomReplicaFromPart(partS string) testlib.ReplicaFunc {
	return func(e *types.Event, c *testlib.Context) (types.ReplicaID, bool) {
		partition, ok := getPartition(c)
		if !ok {
			return "", false
		}
		part, ok := partition.GetPart(partS)
		if !ok {
			return "", false
		}
		return part.ReplicaSet.GetRandom().ID, true
	}
}
