package aigit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

type mockModel struct {
	queryFunc func(ctx context.Context, query string) (string, error)
}

func (m *mockModel) Query(ctx context.Context, query string) (string, error) {
	// Add a small delay to simulate AI processing time
	time.Sleep(50 * time.Millisecond)
	return m.queryFunc(ctx, query)
}

type mockGit struct {
	getStagedDiffFunc    func() (string, error)
	commitFunc           func(message string) error
	getCurrentBranchFunc func() (string, error)
	getBaseBranchFunc    func() (string, error)
	getCommitHistoryFunc func(baseBranch string) (string, error)
	pushFunc             func() error
	forcePushFunc        func() error
	amendFunc            func(message string) error
}

func (m *mockGit) GetStagedDiff() (string, error) {
	return m.getStagedDiffFunc()
}

func (m *mockGit) Commit(message string) error {
	return m.commitFunc(message)
}

func (m *mockGit) GetCurrentBranch() (string, error) {
	return m.getCurrentBranchFunc()
}

func (m *mockGit) GetBaseBranch() (string, error) {
	return m.getBaseBranchFunc()
}

func (m *mockGit) GetCommitHistory(baseBranch string) (string, error) {
	return m.getCommitHistoryFunc(baseBranch)
}

func (m *mockGit) Push() error {
	return m.pushFunc()
}

func (m *mockGit) ForcePush() error {
	return m.forcePushFunc()
}

func (m *mockGit) Amend(message string) error {
	return m.amendFunc(message)
}

type mockGitHub struct {
	createPRFunc           func(title, description string) error
	editPRFunc             func(title, description string) error
	hasOpenPullRequestFunc func() (bool, error)
}

func (m *mockGitHub) CreatePullRequest(title, description string) error {
	return m.createPRFunc(title, description)
}

func (m *mockGitHub) EditPullRequest(title, description string) error {
	return m.editPRFunc(title, description)
}

func (m *mockGitHub) HasOpenPullRequest() (bool, error) {
	return m.hasOpenPullRequestFunc()
}

func TestCli_Commit(t *testing.T) {
	// Create a mock model that returns a predefined commit message
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			return "test: add new feature", nil
		},
	}

	// Create a mock git that returns staged changes and succeeds on commit
	git := &mockGit{
		getStagedDiffFunc: func() (string, error) {
			return "diff --git a/file.txt b/file.txt\n+++ b/file.txt\n@@ -0,0 +1 @@\n+new content", nil
		},
		commitFunc: func(message string) error {
			return nil
		},
	}

	// Create a mock GitHub (not used in commit test but required by NewCli)
	github := &mockGitHub{
		createPRFunc: func(title, description string) error {
			return nil
		},
	}

	cli := NewCli(model, git, github)

	// Test the commit command
	err := cli.Run([]string{"aigit", "commit"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCli_CreatePR(t *testing.T) {
	// Create a mock model that returns a predefined PR description and title
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			if strings.Contains(query, "generate a concise, descriptive title") {
				return "feat: add new feature", nil
			}
			return "This PR adds a new feature that improves the user experience.", nil
		},
	}

	// Create a mock git that returns branch info
	git := &mockGit{
		getCurrentBranchFunc: func() (string, error) {
			return "feature-branch", nil
		},
		getBaseBranchFunc: func() (string, error) {
			return "main", nil
		},
		getCommitHistoryFunc: func(baseBranch string) (string, error) {
			return "abc123 feat: add new feature\ndef456 fix: bug in feature", nil
		},
		pushFunc: func() error {
			return nil
		},
		forcePushFunc: func() error {
			return nil
		},
	}

	// Create a mock GitHub that succeeds on PR creation
	github := &mockGitHub{
		createPRFunc: func(title, description string) error {
			if title != "feat: add new feature" {
				t.Errorf("Expected title 'feat: add new feature', got '%s'", title)
			}
			return nil
		},
		hasOpenPullRequestFunc: func() (bool, error) {
			return false, nil
		},
	}

	cli := NewCli(model, git, github)

	// Test the PR command
	err := cli.Run([]string{"aigit", "pr"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCli_CreatePR_NoCommits(t *testing.T) {
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			return "", nil
		},
	}

	git := &mockGit{
		getCurrentBranchFunc: func() (string, error) {
			return "feature-branch", nil
		},
		getBaseBranchFunc: func() (string, error) {
			return "main", nil
		},
		getCommitHistoryFunc: func(baseBranch string) (string, error) {
			return "", nil
		},
		pushFunc: func() error {
			return nil
		},
		forcePushFunc: func() error {
			return nil
		},
	}

	github := &mockGitHub{
		createPRFunc: func(title, description string) error {
			return nil
		},
	}

	cli := NewCli(model, git, github)

	// Test the PR command with no commits
	err := cli.Run([]string{"aigit", "pr"})
	if err == nil {
		t.Error("Expected error for no commits, got nil")
	}
	if !strings.Contains(err.Error(), "no commits found") {
		t.Errorf("Expected error about no commits, got: %v", err)
	}
}

func TestCli_Commit_WithMarkdown(t *testing.T) {
	// Create a mock model that returns a commit message wrapped in markdown code blocks
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			return "```\ntest: add new feature\n```", nil
		},
	}

	// Create a mock git that returns staged changes and succeeds on commit
	git := &mockGit{
		getStagedDiffFunc: func() (string, error) {
			return "diff --git a/file.txt b/file.txt\n+++ b/file.txt\n@@ -0,0 +1 @@\n+new content", nil
		},
		commitFunc: func(message string) error {
			if message != "test: add new feature" {
				t.Errorf("Expected cleaned message, got: %s", message)
			}
			return nil
		},
	}

	// Create a mock GitHub (not used in commit test but required by NewCli)
	github := &mockGitHub{
		createPRFunc: func(title, description string) error {
			return nil
		},
	}

	cli := NewCli(model, git, github)

	// Test the commit command
	err := cli.Run([]string{"aigit", "commit"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCli_CreatePR_WithMarkdown(t *testing.T) {
	// Create a mock model that returns a PR description and title wrapped in markdown
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			if strings.Contains(query, "generate a concise, descriptive title") {
				return "```\nfeat: add new feature\n```", nil
			}
			return "```\nThis PR adds a new feature that improves the user experience.\n```", nil
		},
	}

	// Create a mock git that returns branch info
	git := &mockGit{
		getCurrentBranchFunc: func() (string, error) {
			return "feature-branch", nil
		},
		getBaseBranchFunc: func() (string, error) {
			return "main", nil
		},
		getCommitHistoryFunc: func(baseBranch string) (string, error) {
			return "abc123 feat: add new feature\ndef456 fix: bug in feature", nil
		},
		pushFunc: func() error {
			return nil
		},
		forcePushFunc: func() error {
			return nil
		},
	}

	// Create a mock GitHub that verifies the cleaned description and title
	github := &mockGitHub{
		createPRFunc: func(title, description string) error {
			expectedTitle := "feat: add new feature"
			expectedDesc := "This PR adds a new feature that improves the user experience."
			if title != expectedTitle {
				t.Errorf("Expected title %q, got %q", expectedTitle, title)
			}
			if description != expectedDesc {
				t.Errorf("Expected description %q, got %q", expectedDesc, description)
			}
			return nil
		},
		hasOpenPullRequestFunc: func() (bool, error) {
			return false, nil
		},
	}

	cli := NewCli(model, git, github)

	// Test the PR command
	err := cli.Run([]string{"aigit", "pr"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCli_CreatePR_GitHubError(t *testing.T) {
	// Create a mock model that returns a predefined PR description
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			if strings.Contains(query, "generate a concise, descriptive title") {
				return "feat: add new feature", nil
			}
			return "This PR adds a new feature that improves the user experience.", nil
		},
	}

	// Create a mock git that returns branch info
	git := &mockGit{
		getCurrentBranchFunc: func() (string, error) {
			return "feature-branch", nil
		},
		getBaseBranchFunc: func() (string, error) {
			return "main", nil
		},
		getCommitHistoryFunc: func(baseBranch string) (string, error) {
			return "abc123 feat: add new feature\ndef456 fix: bug in feature", nil
		},
		pushFunc: func() error {
			return nil
		},
		forcePushFunc: func() error {
			return nil
		},
	}

	// Create a mock GitHub that fails on PR creation
	github := &mockGitHub{
		createPRFunc: func(title, description string) error {
			return fmt.Errorf("gh auth login")
		},
		hasOpenPullRequestFunc: func() (bool, error) {
			return false, nil
		},
		editPRFunc: func(title, description string) error {
			return nil
		},
	}

	cli := NewCli(model, git, github)

	// Test the PR command
	err := cli.Run([]string{"aigit", "pr"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "gh auth login") {
		t.Errorf("Expected error to contain 'gh auth login', got: %v", err)
	}
}

func TestCli_CreatePR_WithForcePush(t *testing.T) {
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			if strings.Contains(query, "generate a concise, descriptive title") {
				return "feat: force push feature", nil
			}
			return "This PR adds a new feature that required force push.", nil
		},
	}
	git := &mockGit{
		getCurrentBranchFunc: func() (string, error) { return "feature-branch", nil },
		getBaseBranchFunc:    func() (string, error) { return "main", nil },
		getCommitHistoryFunc: func(baseBranch string) (string, error) {
			return "abc123 feat: add new feature\ndef456 fix: bug in feature", nil
		},
		pushFunc:      func() error { return fmt.Errorf("push failed") },
		forcePushFunc: func() error { return nil },
	}
	github := &mockGitHub{
		createPRFunc:           func(title, description string) error { return nil },
		hasOpenPullRequestFunc: func() (bool, error) { return false, nil },
		editPRFunc:             func(title, description string) error { return nil },
	}
	cli := NewCli(model, git, github)
	err := cli.Run([]string{"aigit", "pr"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCli_CreatePR_PushFailed(t *testing.T) {
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			return "feat: add new feature\n\nThis PR adds a new feature.", nil
		},
	}

	git := &mockGit{
		getCurrentBranchFunc: func() (string, error) {
			return "feature-branch", nil
		},
		getBaseBranchFunc: func() (string, error) {
			return "main", nil
		},
		getCommitHistoryFunc: func(baseBranch string) (string, error) {
			return "abc123 feat: add new feature", nil
		},
		pushFunc: func() error {
			return fmt.Errorf("push failed")
		},
		forcePushFunc: func() error {
			return fmt.Errorf("force push failed")
		},
	}

	github := &mockGitHub{
		createPRFunc: func(title, description string) error {
			return nil
		},
	}

	cli := NewCli(model, git, github)

	// Test the PR command
	err := cli.Run([]string{"aigit", "pr"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to push branch") {
		t.Errorf("Expected error about push failure, got: %v", err)
	}
}

func TestCleanMarkdownCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "feat: add new feature",
			expected: "feat: add new feature",
		},
		{
			name:     "with markdown",
			input:    "```\nfeat: add new feature\n```",
			expected: "feat: add new feature",
		},
		{
			name:     "with AI prefix",
			input:    "AI: feat: add new feature",
			expected: "feat: add new feature",
		},
		{
			name:     "with AI prefix and markdown",
			input:    "```\nAI: feat: add new feature\n```",
			expected: "feat: add new feature",
		},
		{
			name:     "with duplicate lines",
			input:    "feat: add new feature\n\nfeat: add new feature",
			expected: "feat: add new feature",
		},
		{
			name:     "with AI prefix and duplicate lines",
			input:    "AI: feat: add new feature\n\nfeat: add new feature",
			expected: "feat: add new feature",
		},
		{
			name:     "with AI prefix, markdown, and duplicate lines",
			input:    "```\nAI: feat: add new feature\n\nfeat: add new feature\n```",
			expected: "feat: add new feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanMarkdownCodeBlocks(tt.input)
			if result != tt.expected {
				t.Errorf("cleanMarkdownCodeBlocks(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCli_Amend(t *testing.T) {
	// Create a mock model that returns a predefined commit message
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			return "test: update feature", nil
		},
	}

	// Create a mock git that returns staged changes
	git := &mockGit{
		getStagedDiffFunc: func() (string, error) {
			return "diff --git a/file.txt b/file.txt\nindex abc123..def456 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old line\n+new line", nil
		},
		amendFunc: func(message string) error {
			return nil
		},
	}

	cli := NewCli(model, git, nil)

	// Test the amend command
	err := cli.Run([]string{"aigit", "amend"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCli_Amend_NoChanges(t *testing.T) {
	// Create a mock model
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			return "test: update feature", nil
		},
	}

	// Create a mock git that returns no staged changes
	git := &mockGit{
		getStagedDiffFunc: func() (string, error) {
			return "", nil
		},
	}

	cli := NewCli(model, git, nil)

	// Test the amend command
	err := cli.Run([]string{"aigit", "amend"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no changes staged for amend") {
		t.Errorf("Expected error about no changes, got: %v", err)
	}
}

func TestCli_Amend_WithMarkdown(t *testing.T) {
	// Create a mock model that returns a commit message wrapped in markdown
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			return "```\ntest: update feature\n```", nil
		},
	}

	// Create a mock git that returns staged changes
	git := &mockGit{
		getStagedDiffFunc: func() (string, error) {
			return "diff --git a/file.txt b/file.txt\nindex abc123..def456 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old line\n+new line", nil
		},
		amendFunc: func(message string) error {
			if message != "test: update feature" {
				t.Errorf("Expected message 'test: update feature', got '%s'", message)
			}
			return nil
		},
	}

	cli := NewCli(model, git, nil)

	// Test the amend command
	err := cli.Run([]string{"aigit", "amend"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
