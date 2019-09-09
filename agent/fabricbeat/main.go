package main

import (
	"os"

	"github.com/blockchain-analyzer/agent/fabricbeat/cmd"

	_ "github.com/blockchain-analyzer/agent/fabricbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
