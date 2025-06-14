package aigit

import (
	"context"
	"fmt"
)

type Cli struct {
	model Model
	git   Git
}

func NewCli(model Model, git Git) *Cli {
	return &Cli{
		model: model,
		git:   git,
	}
}

func (cli *Cli) Run(args []string) error {
	// For now, just implement commit command
	if len(args) < 2 || args[1] != "commit" {
		return fmt.Errorf("usage: aigit commit")
	}

	// Get staged changes
	diff, err := cli.git.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("error getting staged changes: %w", err)
	}

	if diff == "" {
		return fmt.Errorf("no changes staged for commit")
	}

	// Ask AI for commit message
	query := fmt.Sprintf("Please write a concise and descriptive commit message for the following changes:\n\n%s", diff)
	message, err := cli.model.Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("error getting commit message from AI: %w", err)
	}

	// Execute git commit
	if err := cli.git.Commit(message); err != nil {
		return fmt.Errorf("error committing changes: %w", err)
	}

	fmt.Printf("Committed with message:\n%s\n", message)
	return nil
}
