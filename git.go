package aigit

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
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
	// Push pushes the current branch to remote
	Push() error
	// ForcePush force pushes the current branch to remote
	ForcePush() error
	// Amend amends the last commit
	Amend(message string) error
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
	return runCommand("git", "diff", "--staged")
}

func (g *GitCli) Commit(message string) error {
	_, err := runCommand("git", "commit", "-m", message)
	return err
}

func (g *GitCli) GetCurrentBranch() (string, error) {
	return runCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
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
	return runCommand("git", "log", "--pretty=format:%h %s", baseBranch+"..HEAD")
}

func (g *GitCli) Push() error {
	output, err := runCommand("git", "push")
	if err != nil && isNoUpstreamError(output) {
		branch, berr := g.GetCurrentBranch()
		if berr != nil {
			return fmt.Errorf("push failed and could not get current branch: %w", berr)
		}
		branch = strings.TrimSpace(branch)
		output, err = runCommand("git", "push", "--set-upstream", "origin", branch)
	}
	return err
}

func (g *GitCli) ForcePush() error {
	output, err := runCommand("git", "push", "--force")
	if err != nil && isNoUpstreamError(output) {
		branch, berr := g.GetCurrentBranch()
		if berr != nil {
			return fmt.Errorf("force push failed and could not get current branch: %w", berr)
		}
		branch = strings.TrimSpace(branch)
		output, err = runCommand("git", "push", "--force", "--set-upstream", "origin", branch)
	}
	return err
}

func (g *GitCli) Amend(message string) error {
	cmd := exec.Command("git", "commit", "--amend", "--no-edit")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error amending commit: %w", err)
	}

	// Update the commit message
	cmd = exec.Command("git", "commit", "--amend", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error updating commit message: %w", err)
	}

	return nil
}

// isNoUpstreamError checks if the output indicates a missing upstream branch
func isNoUpstreamError(output string) bool {
	return strings.Contains(output, "has no upstream branch")
}
