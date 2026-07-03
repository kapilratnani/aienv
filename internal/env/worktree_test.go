package env

import (
	"os"
	"strings"
	"testing"
)

func TestSetupWorktreeNewBranch(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	mounts, cleanup, err := SetupWorktree(&WorktreeConfig{
		RepoPath:   repoDir,
		Branch:     "new-branch",
		BaseBranch: head,
	})
	if err != nil {
		t.Fatalf("SetupWorktree: %v", err)
	}

	if len(mounts) != 2 {
		t.Fatalf("expected 2 mounts, got %d", len(mounts))
	}

	expectedWtPath := WorktreePath(repoDir, "new-branch")
	if mounts[0].Source != expectedWtPath {
		t.Errorf("mounts[0].Source = %q, want %q", mounts[0].Source, expectedWtPath)
	}
	if !mounts[0].Writable {
		t.Error("mounts[0].Writable should be true")
	}

	if !IsGitWorktree(expectedWtPath) {
		t.Error("expected worktree to exist")
	}

	if mounts[1].Writable != true {
		t.Error("mounts[1] (git dir) should be writable")
	}

	if cleanup == nil {
		t.Fatal("expected non-nil cleanup")
	}

	cleanup()
	if IsGitWorktree(expectedWtPath) {
		t.Error("expected worktree to be removed by cleanup")
	}
}

func TestSetupWorktreeKeep(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	_, cleanup, err := SetupWorktree(&WorktreeConfig{
		RepoPath:   repoDir,
		Branch:     "keep-branch",
		BaseBranch: head,
		Keep:       true,
	})
	if err != nil {
		t.Fatalf("SetupWorktree: %v", err)
	}

	if cleanup != nil {
		t.Error("expected nil cleanup when Keep=true")
	}

	expectedWtPath := WorktreePath(repoDir, "keep-branch")
	if !IsGitWorktree(expectedWtPath) {
		t.Error("expected worktree to exist")
	}

	WorktreeRemove(repoDir, expectedWtPath)
}

func TestSetupWorktreeExistingWorktree(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	existingBranch := "existing"
	wtPath := WorktreePath(repoDir, existingBranch)
	run(t, repoDir, "git", "branch", existingBranch, head)
	run(t, repoDir, "git", "worktree", "add", wtPath, existingBranch)

	got, cleanup, err := SetupWorktree(&WorktreeConfig{
		RepoPath:   repoDir,
		Branch:     existingBranch,
		BaseBranch: head,
	})
	if err != nil {
		t.Fatalf("SetupWorktree: %v", err)
	}

	if got[0].Source != wtPath {
		t.Errorf("mounts[0].Source = %q, want %q", got[0].Source, wtPath)
	}

	if cleanup != nil {
		cleanup()
	}
}

func TestSetupWorktreePathCollision(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	wtPath := WorktreePath(repoDir, "collision")
	if err := os.MkdirAll(wtPath, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	_, _, err := SetupWorktree(&WorktreeConfig{
		RepoPath:   repoDir,
		Branch:     "collision",
		BaseBranch: head,
	})
	if err == nil {
		t.Fatal("expected error for non-worktree directory at worktree path")
	}
	if !strings.Contains(err.Error(), "not a git worktree") {
		t.Errorf("unexpected error: %v", err)
	}
}
