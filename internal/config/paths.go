package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func GenerateSessionID() string {
	b := make([]byte, 4)
	rand.Read(b)
	t := time.Now().UTC().Format("20060102-150405")
	return fmt.Sprintf("%s-%s", t, hex.EncodeToString(b))
}

var validName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// test overrides — set in tests to avoid touching real XDG paths.
var (
	TestDataDir  string
	TestTrustDir string
)

func DataDir() string {
	if TestDataDir != "" {
		return TestDataDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "aienv")
}

func EnvDir(name string) string {
	return filepath.Join(DataDir(), name)
}

func EnvYAML(name string) string {
	return filepath.Join(EnvDir(name), "env.yaml")
}

func AuditDir(name, sessionID string) string {
	return filepath.Join(EnvDir(name), "audit", sessionID)
}

func TrustDir() string {
	if TestTrustDir != "" {
		return TestTrustDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "aienv", "trust")
}

func TrustCachePath(hash string) string {
	return filepath.Join(TrustDir(), hash+".json")
}

func IsValidName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid name %q: must be lowercase alphanumeric with hyphens (e.g. backend-api)", name)
	}
	return nil
}

func ComputeHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func ImageTag(hash string) string {
	return fmt.Sprintf("aienv/env:%s", hash)
}

func ListEnvNames() ([]string, error) {
	dir := DataDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", dir, err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && validName.MatchString(e.Name()) {
			path := filepath.Join(dir, e.Name(), "env.yaml")
			if _, err := os.Stat(path); err == nil {
				names = append(names, e.Name())
			}
		}
	}
	return names, nil
}
