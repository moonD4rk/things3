package main

import (
	"os"

	"github.com/moond4rk/things3/cmd/things3/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
