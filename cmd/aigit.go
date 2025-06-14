package main

import (
	"fmt"
	"os"

	"github.com/johanhenriksson/aigit"
)

func main() {
	model := aigit.GetDefaultModel()

	git, err := aigit.NewGit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	github, err := aigit.NewGitHub()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cli := aigit.NewCli(model, git, github)
	if err := cli.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
