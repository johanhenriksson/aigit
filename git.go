package aigit

import (
	"errors"
	"fmt"
	"os/exec"
)

var ErrNoGit = errors.New("git is not installed or not found in PATH")

// Git defines the interface for git operations
type Git interface {
	// GetStagedDiff returns the output of `git diff --staged` command
	GetStagedDiff() (string, error)
	// Commit creates a commit with the given message
	Commit(message string) error
	// GetCurrentBranch returns the name of the current branch
	GetCurrentBranch() (string, error)
	// GetBaseBranch returns the name of the base branch (main/master)
	GetBaseBranch() (string, error)
	// GetCommitHistory returns the commit history between current branch and base branch
	GetCommitHistory(baseBranch string) (string, error)
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

func (g *GitCli) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (g *GitCli) GetBaseBranch() (string, error) {
	// Try main first, then master
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
		if err := cmd.Run(); err == nil {
			return branch, nil
		}
	}
	return "", fmt.Errorf("could not find main or master branch")
}

func (g *GitCli) GetCommitHistory(baseBranch string) (string, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%h %s", baseBranch+"..HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
