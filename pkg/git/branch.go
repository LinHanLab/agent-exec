package git

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// Client provides git operations with event emission
type Client struct {
	emitter events.Emitter
}

// NewClient creates a new git client with the given emitter
func NewClient(emitter events.Emitter) *Client {
	return &Client{emitter: emitter}
}

// RandomBranchName generates a random branch name like "impl-a3f9c2"
func RandomBranchName() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based name if random fails
		return fmt.Sprintf("impl-%d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("impl-%s", hex.EncodeToString(bytes))
}

// CreateBranch creates a new branch from the current HEAD
func (c *Client) CreateBranch(name string) error {
	cmd := exec.Command("git", "checkout", "-b", name)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create branch %s: %s", name, string(output))
	}
	c.emitter.Emit(events.EventGitBranchCreated, events.BranchCreatedData{
		BranchName: name,
		Base:       "", // Empty for CreateBranch
	})
	return nil
}

// CreateBranchFrom creates a new branch from a specified base branch
func (c *Client) CreateBranchFrom(name, base string) error {
	cmd := exec.Command("git", "checkout", "-b", name, base)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create branch %s from %s: %s", name, base, string(output))
	}
	c.emitter.Emit(events.EventGitBranchCreated, events.BranchCreatedData{
		BranchName: name,
		Base:       base, // Set for CreateBranchFrom
	})
	return nil
}

// Checkout switches to the specified branch
func (c *Client) Checkout(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to checkout %s: %s", branch, string(output))
	}
	c.emitter.Emit(events.EventGitBranchCheckedOut, events.BranchCheckedOutData{
		BranchName: branch,
	})
	return nil
}

// SquashCommits squashes all commits on current branch relative to base into one commit
func (c *Client) SquashCommits(base, message string) error {
	// Get the merge base
	mergeBaseCmd := exec.Command("git", "merge-base", base, "HEAD")
	mergeBaseOutput, err := mergeBaseCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find merge base: %w", err)
	}
	mergeBase := strings.TrimSpace(string(mergeBaseOutput))

	// Soft reset to merge base (keeps changes staged)
	resetCmd := exec.Command("git", "reset", "--soft", mergeBase)
	if output, err := resetCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reset to base %s: %s", base, string(output))
	}

	// Stage all changes including untracked files
	addCmd := exec.Command("git", "add", ".")
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stage changes for squash: %s", string(output))
	}

	// Commit all staged changes
	commitCmd := exec.Command("git", "commit", "-m", message)
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit squashed changes: %s", string(output))
	}

	c.emitter.Emit(events.EventGitCommitsSquashed, events.CommitsSquashedData{
		BranchName: base,
	})
	return nil
}

// DeleteBranch deletes the specified branch
func (c *Client) DeleteBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete branch %s: %s", branch, string(output))
	}
	c.emitter.Emit(events.EventGitBranchDeleted, events.BranchDeletedData{
		BranchName: branch,
	})
	return nil
}

// GetCurrentBranch returns the name of the current branch
func (c *Client) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
