package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var validName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func AIEnvsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-envs")
}

func EnvDir(name string) string {
	return filepath.Join(AIEnvsDir(), name)
}

func EnvYAML(name string) string {
	return filepath.Join(EnvDir(name), "ai-env.yaml")
}

func OpenCodeJSON(name string) string {
	return filepath.Join(EnvDir(name), "opencode.json")
}

func AgentConfigPath(name, filename string) string {
	return filepath.Join(EnvDir(name), filename)
}

func IsValidName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid name %q: must be lowercase alphanumeric with hyphens (e.g. backend-api)", name)
	}
	return nil
}

type TrustEntry struct {
	Status     string `json:"status"`
	ReviewedAt string `json:"reviewed_at"`
	EnvName    string `json:"env_name"`
	YAMLHash   string `json:"yaml_hash"`
}

func TrustDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aienv", "trust")
}

func TrustCachePath(hash string) string {
	return filepath.Join(TrustDir(), hash+".json")
}

func ComputeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func ReadTrustCache(hash string) (*TrustEntry, error) {
	path := TrustCachePath(hash)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entry TrustEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func WriteTrustCache(hash, envName string) error {
	dir := TrustDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating trust dir: %w", err)
	}
	entry := TrustEntry{
		Status:     "trusted",
		ReviewedAt: time.Now().UTC().Format(time.RFC3339),
		EnvName:    envName,
		YAMLHash:   hash,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling trust entry: %w", err)
	}
	return os.WriteFile(TrustCachePath(hash), data, 0644)
}

func InvalidateTrustCache(envName string) error {
	dir := TrustDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading trust dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var te TrustEntry
		if err := json.Unmarshal(data, &te); err != nil {
			continue
		}
		if te.EnvName == envName {
			os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
	return nil
}
