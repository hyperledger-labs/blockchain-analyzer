package main

import (
	"os"

	"github.com/balazsprehoda/fabricbeat/cmd"

	_ "github.com/balazsprehoda/fabricbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
