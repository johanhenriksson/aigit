package aigit

import (
	"errors"
	"os/exec"
)

var ErrNoGit = errors.New("git is not installed or not found in PATH")

// Git defines the interface for git operations
type Git interface {
	// GetStagedDiff returns the output of `git diff --staged` command
	GetStagedDiff() (string, error)
	// Commit creates a commit with the given message
	Commit(message string) error
}

// GitCli implements Git interface using actual git commands
type GitCli struct{}

func NewGit() (Git, error) {
	// Verify git is installed
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return nil, errors.Join(ErrNoGit, err)
	}
	return &GitCli{}, nil
}

func (g *GitCli) GetStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (g *GitCli) Commit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	return cmd.Run()
}
