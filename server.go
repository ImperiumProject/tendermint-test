package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ImperiumProject/imperium/config"
	"github.com/ImperiumProject/imperium/testlib"
	"github.com/ImperiumProject/tendermint-test/common"
	"github.com/ImperiumProject/tendermint-test/testcases/rskip"
	"github.com/ImperiumProject/tendermint-test/util"
)

func main() {

	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)

	sysParams := common.NewSystemParams(4)

	server, err := testlib.NewTestingServer(
		&config.Config{
			APIServerAddr: "10.0.0.8:7074",
			NumReplicas:   sysParams.N,
			LogConfig: config.LogConfig{
				Format: "json",
				Path:   "/tmp/tendermint/log/checker.log",
			},
		},
		&util.TMessageParser{},
		[]*testlib.TestCase{
			// testcases.DummyTestCase(),
			rskip.RoundSkip(sysParams, 1, 3),
			// rskip.BlockVotes(sysParams),
			// lockedvalue.DifferentDecisions(sysParams),
			// lockedvalue.Two(),
			// lockedvalue.Three(),
			// sanity.OneTestCase(),
			// sanity.TwoTestCase(),
			// sanity.ThreeTestCase(),
			//sanity.HigherProp(),
			// bfttime.OneTestCase(),
		},
	)

	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err.Error())
		os.Exit(1)
	}

	go func() {
		<-termCh
		server.Stop()
	}()

	server.Start()

}
