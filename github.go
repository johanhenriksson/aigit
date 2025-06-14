package aigit

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var ErrNoGitHubCLI = errors.New("GitHub CLI (gh) is not installed or not found in PATH")

// GitHub defines the interface for GitHub operations
type GitHub interface {
	// CreatePullRequest creates a pull request with the given title and description
	CreatePullRequest(title, description string) error
	// EditPullRequest edits the current pull request with the given title and description
	EditPullRequest(title, description string) error
	// HasOpenPullRequest returns true if there is an open PR for the current branch
	HasOpenPullRequest() (bool, error)
}

// GitHubCLI implements GitHub interface using the GitHub CLI
type GitHubCLI struct{}

func NewGitHub() (GitHub, error) {
	// Verify GitHub CLI is installed
	cmd := exec.Command("gh", "--version")
	if err := cmd.Run(); err != nil {
		return nil, errors.Join(ErrNoGitHubCLI, err)
	}
	return &GitHubCLI{}, nil
}

func (g *GitHubCLI) CreatePullRequest(title, description string) error {
	output, err := runCommand("gh", "pr", "create", "--title", title, "--body", description)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}
	fmt.Print(output)
	return nil
}

func (g *GitHubCLI) EditPullRequest(title, description string) error {
	output, err := runCommand("gh", "pr", "edit", "--title", title, "--body", description)
	if err != nil {
		return fmt.Errorf("failed to edit pull request: %w", err)
	}
	fmt.Print(output)
	return nil
}

func (g *GitHubCLI) HasOpenPullRequest() (bool, error) {
	output, err := runCommand("gh", "pr", "view", "--json", "state", "--jq", ".state")
	if err != nil {
		// If the error is because there is no PR, gh returns a non-zero exit code and output contains "no pull requests found"
		if strings.Contains(output, "no pull requests found") {
			return false, nil
		}
		return false, err
	}
	return strings.TrimSpace(output) == "OPEN", nil
}
