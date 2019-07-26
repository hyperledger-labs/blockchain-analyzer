package main

import (
	"os"

	"github.com/hyperledger-elastic/agent/fabricbeat/cmd"

	_ "github.com/hyperledger-elastic/agent/fabricbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
