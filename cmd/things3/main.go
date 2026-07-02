package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/moond4rk/things3/cmd/things3/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := cmd.NewRootCmd()
	if err := root.ExecuteContext(ctx); err != nil {
		cmd.RenderError(root, err)
		os.Exit(cmd.ExitCode(err))
	}
}
