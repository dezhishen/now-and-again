package main

import (
	"os"

	"github.com/dezhishen/now-and-again/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
