package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
)

func main() {
	// Register subcommands, split and merge
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&splitCmd{}, "")
	subcommands.Register(&mergeCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}

