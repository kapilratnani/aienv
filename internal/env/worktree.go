package env

import (
	"fmt"
	"os"
)

type WorktreeConfig struct {
	RepoPath   string
	Branch     string
	BaseBranch string
	Keep       bool
}

func SetupWorktree(cfg *WorktreeConfig) ([]Mount, func(), error) {
	if err := GitAvailable(); err != nil {
		return nil, nil, fmt.Errorf("git is required for worktree support: %w", err)
	}

	baseBranch := cfg.BaseBranch
	if baseBranch == "" {
		baseBranch = GetDefaultBranch(cfg.RepoPath)
	}

	wtPath := WorktreePath(cfg.RepoPath, cfg.Branch)

	worktreeExists := false
	if info, err := os.Stat(wtPath); err == nil && info.IsDir() {
		if IsGitWorktree(wtPath) {
			worktreeExists = true
		} else {
			return nil, nil, fmt.Errorf("path %q exists but is not a git worktree; remove it manually", wtPath)
		}
	}

	if !worktreeExists {
		if err := WorktreeAdd(cfg.RepoPath, cfg.Branch, baseBranch); err != nil {
			return nil, nil, fmt.Errorf("creating worktree: %w", err)
		}
	}

	gitDir, err := ResolveGitDir(wtPath)
	if err != nil {
		if !worktreeExists {
			WorktreeRemove(cfg.RepoPath, wtPath)
		}
		return nil, nil, fmt.Errorf("resolving git dir: %w", err)
	}

	mounts := []Mount{
		{Source: wtPath, Writable: true},
		{Source: gitDir, Target: gitDir, Writable: true},
	}

	var cleanup func()
	if !cfg.Keep {
		cleanup = func() {
			if err := WorktreeRemove(cfg.RepoPath, wtPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove worktree %s: %v\n", wtPath, err)
			}
		}
	}

	return mounts, cleanup, nil
}
