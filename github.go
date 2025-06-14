package aigit

import (
	"errors"
	"fmt"
	"os/exec"
)

var ErrNoGitHubCLI = errors.New("GitHub CLI (gh) is not installed or not found in PATH")

// GitHub defines the interface for GitHub operations
type GitHub interface {
	// CreatePullRequest creates a pull request with the given title and description
	CreatePullRequest(title, description string) error
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
