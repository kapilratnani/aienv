package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove orphaned Docker images and audit data",
	Long: `Scans ~/.local/share/aienv/ for environments and removes:
  - Docker images with no matching env.yaml (orphaned)
  - Audit directories for deleted environments

Does not touch current environments.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return clean()
	},
}

func clean() error {
	activeNames, err := config.ListEnvNames()
	if err != nil {
		return fmt.Errorf("listing environments: %w", err)
	}

	// Build set of active env YAML hashes
	activeHashes := make(map[string]bool)
	for _, name := range activeNames {
		yamlPath := config.EnvYAML(name)
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: %s: %v\n", yamlPath, err)
			continue
		}
		activeHashes[config.ComputeHash(data)] = true
	}

	// List all aienv Docker images
	out, err := exec.Command("docker", "images", "--filter=reference=aienv/env:*", "--format={{.Tag}}").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: listing Docker images: %v\n", err)
	} else {
		orphaned := 0
		for _, tag := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if !activeHashes[tag] {
				exec.Command("docker", "rmi", "-f", fmt.Sprintf("aienv/env:%s", tag)).Run()
				orphaned++
			}
		}
		if orphaned > 0 {
			fmt.Fprintf(os.Stderr, "Removed %d orphaned image(s).\n", orphaned)
		} else {
			fmt.Fprintf(os.Stderr, "No orphaned images found.\n")
		}
	}

	// Remove orphaned audit dirs (for deleted envs)
	dataDir := config.DataDir()
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "No data directory found.\n")
			return nil
		}
		return fmt.Errorf("reading %s: %w", dataDir, err)
	}

	activeSet := make(map[string]bool)
	for _, n := range activeNames {
		activeSet[n] = true
	}

	cleaned := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		auditDir := filepath.Join(dataDir, e.Name(), "audit")
		if !activeSet[e.Name()] {
			// Entire env dir is orphaned (shouldn't happen since we already list active envs, but just in case)
			if info, err := os.Stat(auditDir); err == nil && info.IsDir() {
				os.RemoveAll(auditDir)
				cleaned++
			}
		}
	}

	// Also clean up trust cache for stale/gone envs
	trustDir := config.TrustDir()
	if trustEntries, err := os.ReadDir(trustDir); err == nil {
		for _, te := range trustEntries {
			if !te.IsDir() {
				os.Remove(filepath.Join(trustDir, te.Name()))
			}
		}
		fmt.Fprintf(os.Stderr, "Cleaned trust cache.\n")
	}

	if cleaned > 0 {
		fmt.Fprintf(os.Stderr, "Removed %d orphaned audit dir(s).\n", cleaned)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
