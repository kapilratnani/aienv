package env

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GitAvailable() error {
	cmd := exec.Command("git", "version")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git not found: %w\n%s", err, out)
	}
	return nil
}

func RepoName(repoPath string) string {
	return filepath.Base(repoPath)
}

func WorktreePath(repoPath, branch string) string {
	parent := filepath.Dir(repoPath)
	name := RepoName(repoPath)
	return filepath.Join(parent, name+"-"+branch)
}

func GetDefaultBranch(repoPath string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		if strings.HasPrefix(ref, "refs/remotes/origin/") {
			return ref[20:]
		}
	}
	for _, candidate := range []string{"main", "master"} {
		cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+candidate)
		cmd.Dir = repoPath
		if cmd.Run() == nil {
			return candidate
		}
	}
	return ""
}

func BranchExistsOnRemote(repoPath, branch string) bool {
	cmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

func BranchExistsLocally(repoPath, branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch)
	cmd.Dir = repoPath
	return cmd.Run() == nil
}

func WorktreeAdd(repoPath, branch, baseBranch string) error {
	existsLocal := BranchExistsLocally(repoPath, branch)
	existsRemote := BranchExistsOnRemote(repoPath, branch)

	var args []string
	wtPath := WorktreePath(repoPath, branch)

	switch {
	case existsLocal:
		args = []string{"worktree", "add", wtPath, branch}
	case existsRemote:
		fetchCmd := exec.Command("git", "fetch", "origin", branch)
		fetchCmd.Dir = repoPath
		if out, err := fetchCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git fetch origin %s: %w\n%s", branch, err, out)
		}
		branchCmd := exec.Command("git", "branch", branch, "origin/"+branch)
		branchCmd.Dir = repoPath
		if out, err := branchCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git branch %s origin/%s: %w\n%s", branch, branch, err, out)
		}
		args = []string{"worktree", "add", wtPath, branch}
	default:
		if baseBranch == "" {
			baseBranch = GetDefaultBranch(repoPath)
		}
		if baseBranch == "" {
			return fmt.Errorf("cannot determine base branch; specify with --worktree-base")
		}
		args = []string{"worktree", "add", "-b", branch, wtPath, baseBranch}
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v: %w\n%s", args, err, out)
	}
	return nil
}

func WorktreeRemove(repoPath, worktreePath string) error {
	cmd := exec.Command("git", "worktree", "remove", "--force", worktreePath)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("removing worktree %s: %w\n%s", worktreePath, err, out)
	}
	return nil
}

func IsGitWorktree(path string) bool {
	fi, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func ResolveGitDir(worktreePath string) (string, error) {
	data, err := os.ReadFile(filepath.Join(worktreePath, ".git"))
	if err != nil {
		return "", fmt.Errorf("reading .git in worktree: %w", err)
	}
	content := strings.TrimSpace(string(data))
	prefix := "gitdir: "
	if !strings.HasPrefix(content, prefix) {
		return "", fmt.Errorf("invalid .git file: expected 'gitdir: ...' prefix")
	}
	gitDir := content[len(prefix):]
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(worktreePath, gitDir)
	}
	return filepath.Clean(gitDir), nil
}
