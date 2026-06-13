package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kapilratnani/aienv/internal/config"
	"github.com/kapilratnani/aienv/internal/env"
)

type TrustRecord struct {
	EnvName    string `json:"env_name"`
	Status     string `json:"status"`
	ReviewedAt string `json:"reviewed_at"`
}

func NeedsTrustPrompt(e *env.Env) (bool, error) {
	yamlPath := config.EnvYAML(e.Meta.Name)
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return false, fmt.Errorf("reading env.yaml: %w", err)
	}

	hash := config.ComputeHash(data)
	cachePath := config.TrustCachePath(hash)

	if _, err := os.Stat(cachePath); err == nil {
		var record TrustRecord
		raw, _ := os.ReadFile(cachePath)
		if json.Unmarshal(raw, &record) == nil && record.Status == "trusted" {
			return false, nil
		}
	}

	return true, nil
}

func TrustPrompt(e *env.Env) error {
	fmt.Fprintf(os.Stderr, "\n--- Environment Trust Review ---\n")
	fmt.Fprintf(os.Stderr, "  Name:    %s\n", e.Meta.Name)
	if e.Meta.Description != "" {
		fmt.Fprintf(os.Stderr, "  Description: %s\n", e.Meta.Description)
	}
	fmt.Fprintf(os.Stderr, "  Agent:   %s\n", e.Agent.Command)

	fmt.Fprintf(os.Stderr, "\n  Mounts:\n")
	for _, m := range e.Agent.Mounts {
		rw := "ro"
		if m.Writable {
			rw = "rw"
		}
		fmt.Fprintf(os.Stderr, "    %s → %s (%s)\n", m.Source, m.Target, rw)
	}

	if e.Permissions != nil && e.Permissions.Network != nil {
		fmt.Fprintf(os.Stderr, "\n  Network:\n")
		for _, h := range e.Permissions.Network.Allow {
			fmt.Fprintf(os.Stderr, "    allow: %s\n", h)
		}
		for _, h := range e.Permissions.Network.Deny {
			fmt.Fprintf(os.Stderr, "    deny: %s\n", h)
		}
	}

	if e.Audit.Persist {
		fmt.Fprintf(os.Stderr, "\n  Audit logging enabled\n")
	}

	fmt.Fprintf(os.Stderr, "\nTrust this environment? (y/N): ")

	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		return fmt.Errorf("trust rejected")
	}

	return nil
}

func SaveTrust(e *env.Env) error {
	yamlPath := config.EnvYAML(e.Meta.Name)
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("reading env.yaml: %w", err)
	}

	hash := config.ComputeHash(data)
	cachePath := config.TrustCachePath(hash)

	record := TrustRecord{
		EnvName:    e.Meta.Name,
		Status:     "trusted",
		ReviewedAt: time.Now().UTC().Format(time.RFC3339),
	}

	raw, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling trust record: %w", err)
	}

	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating trust dir: %w", err)
	}

	return os.WriteFile(cachePath, raw, 0644)
}
