package main

import (
	"os"

	"github.com/jf-ferraz/mind-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
