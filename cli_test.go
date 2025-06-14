package aigit

import (
	"context"
	"strings"
	"testing"
)

type mockModel struct {
	queryFunc func(ctx context.Context, query string) (string, error)
}

func (m *mockModel) Query(ctx context.Context, query string) (string, error) {
	return m.queryFunc(ctx, query)
}

type mockGit struct {
	getStagedDiffFunc    func() (string, error)
	commitFunc           func(message string) error
	getCurrentBranchFunc func() (string, error)
	getBaseBranchFunc    func() (string, error)
	getCommitHistoryFunc func(baseBranch string) (string, error)
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

type mockGitHub struct {
	createPRFunc func(title, description string) error
}

func (m *mockGitHub) CreatePullRequest(title, description string) error {
	return m.createPRFunc(title, description)
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
	// Create a mock model that returns a predefined PR description
	model := &mockModel{
		queryFunc: func(ctx context.Context, query string) (string, error) {
			return "feat: add new feature\n\nThis PR adds a new feature that improves the user experience.", nil
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
	}

	// Create a mock GitHub that succeeds on PR creation
	github := &mockGitHub{
		createPRFunc: func(title, description string) error {
			return nil
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
