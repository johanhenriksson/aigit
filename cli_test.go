package aigit

import (
	"context"
	"testing"
)

type mockModel struct {
	queryFunc func(ctx context.Context, query string) (string, error)
}

func (m *mockModel) Query(ctx context.Context, query string) (string, error) {
	return m.queryFunc(ctx, query)
}

type mockGit struct {
	getStagedDiffFunc func() (string, error)
	commitFunc        func(message string) error
}

func (m *mockGit) GetStagedDiff() (string, error) {
	return m.getStagedDiffFunc()
}

func (m *mockGit) Commit(message string) error {
	return m.commitFunc(message)
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

	cli := NewCli(model, git)

	// Test the commit command
	err := cli.Run([]string{"aigit", "commit"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
