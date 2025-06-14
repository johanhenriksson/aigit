package aigit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Cli struct {
	model Model
}

func NewCli(model Model) *Cli {
	return &Cli{
		model: model,
	}
}

func (cli *Cli) Run() {
	// For now, just implement commit command
	if len(os.Args) < 2 || os.Args[1] != "commit" {
		fmt.Println("Usage: aigit commit")
		os.Exit(1)
	}

	// Get staged changes
	diff, err := GetStagedDiff()
	if err != nil {
		fmt.Printf("Error getting staged changes: %v\n", err)
		os.Exit(1)
	}

	if diff == "" {
		fmt.Println("No changes staged for commit")
		os.Exit(1)
	}

	// Ask AI for commit message
	query := fmt.Sprintf("Please write a concise and descriptive commit message for the following changes:\n\n%s", diff)
	message, err := cli.model.Query(context.Background(), query)
	if err != nil {
		fmt.Printf("Error getting commit message from AI: %v\n", err)
		os.Exit(1)
	}

	// Execute git commit
	cmd := exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error committing changes: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Committed with message:\n%s\n", message)
}
