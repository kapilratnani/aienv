package env

import (
	"os/exec"
	"strings"
	"testing"
)

func defaultBranch(t *testing.T, dir string) string {
	out := run(t, dir, "git", "symbolic-ref", "--short", "HEAD")
	return strings.TrimSpace(out)
}

func initRepo(t *testing.T, dir string) string {
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "init")
	return defaultBranch(t, dir)
}

func run(t *testing.T, dir, name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return string(out)
}

func addWorktree(t *testing.T, repoDir, branch, base string) string {
	run(t, repoDir, "git", "branch", branch, base)
	wtDir := t.TempDir()
	run(t, repoDir, "git", "worktree", "add", wtDir, branch)
	return wtDir
}

func TestIsGitWorktree(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	wtDir := addWorktree(t, repoDir, "test-branch", head)

	if !IsGitWorktree(wtDir) {
		t.Error("expected worktree to be detected as git worktree")
	}
	if IsGitWorktree(repoDir) {
		t.Error("expected main repo to NOT be detected as git worktree")
	}
}

func TestResolveGitDir(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	wtDir := addWorktree(t, repoDir, "test-branch", head)

	if _, err := ResolveGitDir(wtDir); err != nil {
		t.Fatalf("ResolveGitDir: %v", err)
	}
}

func TestRepoName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/home/user/projects/myrepo", "myrepo"},
		{"/home/user/projects/my-repo", "my-repo"},
		{"/home/user/projects/repo.git", "repo.git"},
		{"relative/path/to/repo", "repo"},
	}
	for _, tt := range tests {
		got := RepoName(tt.path)
		if got != tt.want {
			t.Errorf("RepoName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestWorktreePath(t *testing.T) {
	tests := []struct {
		repoPath string
		branch   string
		want     string
	}{
		{"/home/user/projects/myrepo", "feature-x", "/home/user/projects/myrepo-feature-x"},
		{"/home/user/projects/my-repo", "bugfix/y", "/home/user/projects/my-repo-bugfix/y"},
	}
	for _, tt := range tests {
		got := WorktreePath(tt.repoPath, tt.branch)
		if got != tt.want {
			t.Errorf("WorktreePath(%q, %q) = %q, want %q", tt.repoPath, tt.branch, got, tt.want)
		}
	}
}

func TestBranchExistsLocally(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	if !BranchExistsLocally(repoDir, head) {
		t.Errorf("expected %q branch to exist locally", head)
	}
	if BranchExistsLocally(repoDir, "nonexistent") {
		t.Error("expected 'nonexistent' branch to not exist locally")
	}
}

func TestBranchExistsOnRemote(t *testing.T) {
	repoDir := t.TempDir()
	initRepo(t, repoDir)

	if BranchExistsOnRemote(repoDir, "main") {
		t.Error("expected false when no remote is configured")
	}
}

func TestGetDefaultBranch(t *testing.T) {
	repoDir := t.TempDir()
	initRepo(t, repoDir)

	branch := GetDefaultBranch(repoDir)
	if branch == "" {
		t.Error("expected GetDefaultBranch to find a branch")
	}
}

func TestGitAvailable(t *testing.T) {
	if err := GitAvailable(); err != nil {
		t.Fatalf("GitAvailable() failed: %v", err)
	}
}

func TestWorktreeAddAndRemove(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	if err := WorktreeAdd(repoDir, "test-worktree", head); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	wtPath := WorktreePath(repoDir, "test-worktree")

	if !IsGitWorktree(wtPath) {
		t.Fatal("expected worktree to exist after WorktreeAdd")
	}

	if err := WorktreeRemove(repoDir, wtPath); err != nil {
		t.Fatalf("WorktreeRemove: %v", err)
	}

	if IsGitWorktree(wtPath) {
		t.Error("expected worktree to be removed")
	}
}

func TestWorktreeAddRemoteOnly(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	remoteDir := t.TempDir()
	run(t, remoteDir, "git", "init", "--bare")

	run(t, repoDir, "git", "remote", "add", "origin", remoteDir)
	run(t, repoDir, "git", "push", "-u", "origin", head)

	run(t, repoDir, "git", "checkout", "-b", "feature-remote")
	run(t, repoDir, "git", "commit", "--allow-empty", "-m", "feature work")
	run(t, repoDir, "git", "push", "-u", "origin", "feature-remote")

	run(t, repoDir, "git", "checkout", head)
	run(t, repoDir, "git", "branch", "-D", "feature-remote")

	if err := WorktreeAdd(repoDir, "feature-remote", ""); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	wtPath := WorktreePath(repoDir, "feature-remote")
	if !IsGitWorktree(wtPath) {
		t.Error("expected worktree to exist")
	}

	WorktreeRemove(repoDir, wtPath)
}

func TestWorktreeAddNewBranch(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	branchName := "brand-new-branch"
	if err := WorktreeAdd(repoDir, branchName, head); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	if !BranchExistsLocally(repoDir, branchName) {
		t.Error("expected new branch to exist locally")
	}

	wtPath := WorktreePath(repoDir, branchName)
	if !IsGitWorktree(wtPath) {
		t.Error("expected worktree to exist")
	}

	WorktreeRemove(repoDir, wtPath)
}

func TestWorktreeAddRemoteOnly_NoFetchNeeded(t *testing.T) {
	repoDir := t.TempDir()
	head := initRepo(t, repoDir)

	remoteDir := t.TempDir()
	run(t, remoteDir, "git", "init", "--bare")

	run(t, repoDir, "git", "remote", "add", "origin", remoteDir)
	run(t, repoDir, "git", "push", "-u", "origin", head)

	run(t, repoDir, "git", "checkout", "-b", "feature-prefetched")
	run(t, repoDir, "git", "commit", "--allow-empty", "-m", "feature")
	run(t, repoDir, "git", "push", "-u", "origin", "feature-prefetched")

	run(t, repoDir, "git", "checkout", head)
	run(t, repoDir, "git", "branch", "-D", "feature-prefetched")
	run(t, repoDir, "git", "fetch", "origin")

	if BranchExistsLocally(repoDir, "feature-prefetched") {
		t.Fatal("precondition failed: branch should not exist locally")
	}

	if !BranchExistsOnRemote(repoDir, "feature-prefetched") {
		t.Fatal("precondition failed: branch should exist on remote")
	}

	if err := WorktreeAdd(repoDir, "feature-prefetched", ""); err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	wtPath := WorktreePath(repoDir, "feature-prefetched")
	if !IsGitWorktree(wtPath) {
		t.Error("expected worktree to exist")
	}

	WorktreeRemove(repoDir, wtPath)
}
