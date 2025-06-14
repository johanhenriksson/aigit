package aigit

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("CLI", func() {
	var (
		model  *mockModel
		git    *mockGit
		github *mockGitHub
		cli    *Cli
	)

	BeforeEach(func() {
		model = &mockModel{}
		git = &mockGit{}
		github = &mockGitHub{}
		cli = NewCli(model, git, github)
	})

	Describe("Commit", func() {
		Context("when there are staged changes", func() {
			BeforeEach(func() {
				model.queryFunc = func(ctx context.Context, query string) (string, error) {
					time.Sleep(50 * time.Millisecond)
					return "test: add new feature", nil
				}
				git.getStagedDiffFunc = func() (string, error) {
					return "diff --git a/file.txt b/file.txt\n+++ b/file.txt\n@@ -0,0 +1 @@\n+new content", nil
				}
				git.commitFunc = func(message string) error {
					return nil
				}
				github.createPRFunc = func(title, description string) error {
					return nil
				}
			})

			It("should create a commit with AI-generated message", func() {
				err := cli.Run([]string{"aigit", "commit"})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when there are no staged changes", func() {
			BeforeEach(func() {
				git.getStagedDiffFunc = func() (string, error) {
					return "", nil
				}
			})

			It("should return an error", func() {
				err := cli.Run([]string{"aigit", "commit"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no changes staged for commit"))
			})
		})

		Context("when the AI response contains markdown", func() {
			BeforeEach(func() {
				model.queryFunc = func(ctx context.Context, query string) (string, error) {
					time.Sleep(50 * time.Millisecond)
					return "```\ntest: add new feature\n```", nil
				}
				git.getStagedDiffFunc = func() (string, error) {
					return "diff --git a/file.txt b/file.txt\n+++ b/file.txt\n@@ -0,0 +1 @@\n+new content", nil
				}
				git.commitFunc = func(message string) error {
					Expect(message).To(Equal("test: add new feature"))
					return nil
				}
				github.createPRFunc = func(title, description string) error {
					return nil
				}
			})

			It("should clean the markdown from the commit message", func() {
				err := cli.Run([]string{"aigit", "commit"})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CreatePR", func() {
		Context("when there are commits to create a PR for", func() {
			BeforeEach(func() {
				model.queryFunc = func(ctx context.Context, query string) (string, error) {
					time.Sleep(50 * time.Millisecond)
					if strings.Contains(query, "generate a concise, descriptive title") {
						return "feat: add new feature", nil
					}
					return "This PR adds a new feature that improves the user experience.", nil
				}
				git.getCurrentBranchFunc = func() (string, error) {
					return "feature-branch", nil
				}
				git.getBaseBranchFunc = func() (string, error) {
					return "main", nil
				}
				git.getCommitHistoryFunc = func(baseBranch string) (string, error) {
					return "abc123 feat: add new feature\ndef456 fix: bug in feature", nil
				}
				git.pushFunc = func() error {
					return nil
				}
				git.forcePushFunc = func() error {
					return nil
				}
				github.createPRFunc = func(title, description string) error {
					Expect(title).To(Equal("feat: add new feature"))
					return nil
				}
				github.hasOpenPullRequestFunc = func() (bool, error) {
					return false, nil
				}
			})

			It("should create a pull request", func() {
				err := cli.Run([]string{"aigit", "pr"})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when there are no commits", func() {
			BeforeEach(func() {
				git.getCurrentBranchFunc = func() (string, error) {
					return "feature-branch", nil
				}
				git.getBaseBranchFunc = func() (string, error) {
					return "main", nil
				}
				git.getCommitHistoryFunc = func(baseBranch string) (string, error) {
					return "", nil
				}
			})

			It("should return an error", func() {
				err := cli.Run([]string{"aigit", "pr"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no commits found"))
			})
		})

		Context("when GitHub authentication fails", func() {
			BeforeEach(func() {
				model.queryFunc = func(ctx context.Context, query string) (string, error) {
					time.Sleep(50 * time.Millisecond)
					if strings.Contains(query, "generate a concise, descriptive title") {
						return "feat: add new feature", nil
					}
					return "This PR adds a new feature.", nil
				}
				git.getCurrentBranchFunc = func() (string, error) {
					return "feature-branch", nil
				}
				git.getBaseBranchFunc = func() (string, error) {
					return "main", nil
				}
				git.getCommitHistoryFunc = func(baseBranch string) (string, error) {
					return "abc123 feat: add new feature", nil
				}
				git.pushFunc = func() error {
					return nil
				}
				git.forcePushFunc = func() error {
					return nil
				}
				github.createPRFunc = func(title, description string) error {
					return fmt.Errorf("gh auth login")
				}
				github.hasOpenPullRequestFunc = func() (bool, error) {
					return false, nil
				}
				github.editPRFunc = func(title, description string) error {
					return nil
				}
			})

			It("should return an error", func() {
				err := cli.Run([]string{"aigit", "pr"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("gh auth login"))
			})
		})

		Context("when force push is required", func() {
			BeforeEach(func() {
				model.queryFunc = func(ctx context.Context, query string) (string, error) {
					time.Sleep(50 * time.Millisecond)
					if strings.Contains(query, "generate a concise, descriptive title") {
						return "feat: force push feature", nil
					}
					return "This PR adds a new feature that required force push.", nil
				}
				git.getCurrentBranchFunc = func() (string, error) {
					return "feature-branch", nil
				}
				git.getBaseBranchFunc = func() (string, error) {
					return "main", nil
				}
				git.getCommitHistoryFunc = func(baseBranch string) (string, error) {
					return "abc123 feat: add new feature\ndef456 fix: bug in feature", nil
				}
				git.pushFunc = func() error {
					return fmt.Errorf("push failed")
				}
				git.forcePushFunc = func() error {
					return nil
				}
				github.createPRFunc = func(title, description string) error {
					return nil
				}
				github.hasOpenPullRequestFunc = func() (bool, error) {
					return false, nil
				}
			})

			It("should attempt force push and create PR", func() {
				err := cli.Run([]string{"aigit", "pr"})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when both push and force push fail", func() {
			BeforeEach(func() {
				model.queryFunc = func(ctx context.Context, query string) (string, error) {
					time.Sleep(50 * time.Millisecond)
					return "feat: add new feature", nil
				}
				git.getCurrentBranchFunc = func() (string, error) {
					return "feature-branch", nil
				}
				git.getBaseBranchFunc = func() (string, error) {
					return "main", nil
				}
				git.getCommitHistoryFunc = func(baseBranch string) (string, error) {
					return "abc123 feat: add new feature", nil
				}
				git.pushFunc = func() error {
					return fmt.Errorf("push failed")
				}
				git.forcePushFunc = func() error {
					return fmt.Errorf("force push failed")
				}
				github.createPRFunc = func(title, description string) error {
					return nil
				}
			})

			It("should return an error", func() {
				err := cli.Run([]string{"aigit", "pr"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to push branch"))
			})
		})
	})

	Describe("Amend", func() {
		Context("when there are staged changes", func() {
			BeforeEach(func() {
				model.queryFunc = func(ctx context.Context, query string) (string, error) {
					time.Sleep(50 * time.Millisecond)
					return "test: update feature", nil
				}
				git.getStagedDiffFunc = func() (string, error) {
					return "diff --git a/file.txt b/file.txt\nindex abc123..def456 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old line\n+new line", nil
				}
				git.amendFunc = func(message string) error {
					return nil
				}
			})

			It("should amend the commit with AI-generated message", func() {
				err := cli.Run([]string{"aigit", "amend"})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when there are no staged changes", func() {
			BeforeEach(func() {
				git.getStagedDiffFunc = func() (string, error) {
					return "", nil
				}
			})

			It("should return an error", func() {
				err := cli.Run([]string{"aigit", "amend"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no changes staged for amend"))
			})
		})

		Context("when the AI response contains markdown", func() {
			BeforeEach(func() {
				model.queryFunc = func(ctx context.Context, query string) (string, error) {
					time.Sleep(50 * time.Millisecond)
					return "```\ntest: update feature\n```", nil
				}
				git.getStagedDiffFunc = func() (string, error) {
					return "diff --git a/file.txt b/file.txt\nindex abc123..def456 100644\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old line\n+new line", nil
				}
				git.amendFunc = func(message string) error {
					Expect(message).To(Equal("test: update feature"))
					return nil
				}
			})

			It("should clean the markdown from the commit message", func() {
				err := cli.Run([]string{"aigit", "amend"})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CleanMarkdownCodeBlocks", func() {
		DescribeTable("cleaning markdown and AI prefixes",
			func(input, expected string) {
				result := cleanMarkdownCodeBlocks(input)
				Expect(result).To(Equal(expected))
			},
			Entry("simple text", "feat: add new feature", "feat: add new feature"),
			Entry("with markdown", "```\nfeat: add new feature\n```", "feat: add new feature"),
			Entry("with AI prefix", "AI: feat: add new feature", "feat: add new feature"),
			Entry("with AI prefix and markdown", "```\nAI: feat: add new feature\n```", "feat: add new feature"),
			Entry("with duplicate lines", "feat: add new feature\n\nfeat: add new feature", "feat: add new feature"),
			Entry("with AI prefix and duplicate lines", "AI: feat: add new feature\n\nfeat: add new feature", "feat: add new feature"),
			Entry("with AI prefix, markdown, and duplicate lines", "```\nAI: feat: add new feature\n\nfeat: add new feature\n```", "feat: add new feature"),
		)
	})
})
