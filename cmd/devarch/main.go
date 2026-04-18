package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type exitCoder interface {
	ExitCode() int
}

type silentError interface {
	Silent() bool
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr, defaultServiceFactory); err != nil {
		if silent, ok := err.(silentError); !ok || !silent.Silent() {
			fmt.Fprintln(os.Stderr, err)
		}
		if coded, ok := err.(exitCoder); ok {
			os.Exit(coded.ExitCode())
		}
		os.Exit(1)
	}
}
