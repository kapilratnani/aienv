package env

import (
	"encoding/json"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

type modelsEntry struct {
	API  string `json:"api"`
	Name string `json:"name"`
	NPM  string `json:"npm,omitempty"`
}

var sdkDefaults = map[string]string{
	"anthropic":  "https://api.anthropic.com",
	"openai":     "https://api.openai.com/v1",
	"groq":       "https://api.groq.com/openai/v1",
	"google":     "https://generativelanguage.googleapis.com/v1beta",
	"mistral":    "https://api.mistral.ai/v1",
	"cohere":     "https://api.cohere.ai/v1",
	"perplexity": "https://api.perplexity.ai",
	"togetherai": "https://api.together.xyz/v1",
	"cerebras":   "https://api.cerebras.ai/v1",
	"deepinfra":  "https://api.deepinfra.com/v1/openai",
	"xai":        "https://api.x.ai/v1",
}

func DetectProviderEndpoints(agent string) []string {
	unique := map[string]bool{}
	var hosts []string

	addHost := func(rawURL string) {
		u, err := url.Parse(rawURL)
		if err != nil {
			return
		}
		host := u.Hostname()
		if host == "" || isLoopback(host) {
			slog.Debug("skipping loopback host", "host", host)
			return
		}
		if !unique[host] {
			unique[host] = true
			hosts = append(hosts, host)
		}
	}

	switch agent {
	case "opencode":
		detectOpenCodeProviders(addHost)
	case "claude-code":
		slog.Debug("adding default Claude endpoint")
		addHost("https://api.anthropic.com")
	}

	for _, envKey := range []string{"OPENAI_BASE_URL", "ANTHROPIC_BASE_URL", "AZURE_OPENAI_ENDPOINT"} {
		if val := os.Getenv(envKey); val != "" {
			slog.Debug("adding from env var", "key", envKey)
			addHost(val)
		}
	}

	if len(hosts) > 0 {
		slog.Debug("provider endpoints detected", "hosts", hosts)
	}
	return hosts
}

func detectOpenCodeProviders(addHost func(string)) {
	models := loadModelsCache()
	displayNames := parseProviderList()

	addCustomConfigProviders(addHost)

	slog.Debug("parsed provider display names", "names", displayNames)

	for _, display := range displayNames {
		providerID := findProviderID(models, display)
		if providerID == "" {
			slog.Debug("no match for display name", "display", display)
			continue
		}
		entry, ok := models[providerID]
		if ok && entry.API != "" {
			slog.Debug("mapped provider to API URL", "display", display, "id", providerID, "url", entry.API)
			addHost(entry.API)
			continue
		}
		if url, ok := sdkDefaults[providerID]; ok {
			slog.Debug("using SDK default for provider", "display", display, "id", providerID, "url", url)
			addHost(url)
		} else {
			slog.Debug("no URL found for provider", "display", display, "id", providerID)
		}
	}
}

func loadModelsCache() map[string]modelsEntry {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".cache", "opencode", "models.json")
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Debug("models.json not found", "error", err)
		return nil
	}
	var models map[string]modelsEntry
	if err := json.Unmarshal(data, &models); err != nil {
		slog.Debug("models.json parse error", "error", err)
		return nil
	}
	slog.Debug("loaded models.json", "providers", len(models))
	return models
}

func parseProviderList() []string {
	cmd := exec.Command("opencode", "providers", "list")
	slog.Debug("running command", "args", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		slog.Debug("command failed", "error", err)
		return nil
	}

	clean := ansiRegex.ReplaceAllString(string(out), "")
	slog.Debug("raw provider list output", "output", clean)

	var names []string
	for line := range strings.SplitSeq(clean, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "●") {
			continue
		}
		_, after, _ := strings.Cut(line, "●")
		rest := after
		rest = strings.TrimSpace(rest)
		if rest == "" {
			continue
		}
		parts := strings.Fields(rest)
		if len(parts) <= 1 {
			continue
		}
		name := strings.Join(parts[:len(parts)-1], " ")
		slog.Debug("found provider", "name", name)
		names = append(names, name)
	}
	return names
}

func findProviderID(models map[string]modelsEntry, display string) string {
	for id, entry := range models {
		if entry.Name == display {
			return id
		}
	}
	lower := strings.ToLower(strings.ReplaceAll(display, " ", "-"))
	if _, ok := models[lower]; ok {
		return lower
	}
	return ""
}

func addCustomConfigProviders(addHost func(string)) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "opencode", "opencode.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg struct {
		Providers map[string]struct {
			Options struct {
				BaseURL string `json:"baseURL"`
			} `json:"options"`
		} `json:"provider"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	for _, p := range cfg.Providers {
		if p.Options.BaseURL != "" {
			addHost(p.Options.BaseURL)
		}
	}
}

func isLoopback(host string) bool {
	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "0.0.0.0" ||
		host == "::1" ||
		strings.HasPrefix(host, "127.")
}
