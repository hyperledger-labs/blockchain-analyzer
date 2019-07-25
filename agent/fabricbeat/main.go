package main

import (
	"os"

	"hyperledger-elastic/agent/fabricbeat/cmd"

	_ "hyperledger-elastic/agent/fabricbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
