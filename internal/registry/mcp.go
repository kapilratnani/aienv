package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OfficialRegistry struct {
	client *http.Client
}

type registryResponse struct {
	Servers []struct {
		Server struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Packages    []struct {
				RegistryType string `json:"registryType"`
				Identifier   string `json:"identifier"`
				Transport    struct {
					Type string `json:"type"`
				} `json:"transport"`
			} `json:"packages"`
		} `json:"server"`
	} `json:"servers"`
	Metadata struct {
		Count int `json:"count"`
	} `json:"metadata"`
}

func NewOfficialRegistry() *OfficialRegistry {
	return &OfficialRegistry{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *OfficialRegistry) Search(ctx context.Context, query string, limit int) ([]MCPServerItem, error) {
	u, _ := url.Parse("https://registry.modelcontextprotocol.io/v0.1/servers")
	q := u.Query()
	q.Set("search", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("version", "latest")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	var regResp registryResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	items := make([]MCPServerItem, 0, len(regResp.Servers))
	for _, s := range regResp.Servers {
		shortName := shortRegistryName(s.Server.Name)
		item := MCPServerItem{
			Name:        shortName,
			DisplayName: s.Server.Name,
			Description: s.Server.Description,
			Type:        "local",
			Source:      "registry",
		}
		if len(s.Server.Packages) > 0 {
			pkg := s.Server.Packages[0]
			item.RegistryType = pkg.RegistryType
			item.Package = pkg.Identifier
			if pkg.Transport.Type == "stdio" {
				item.Type = "local"
			}
			item.Command = registryCommand(pkg.RegistryType, pkg.Identifier)
		}
		items = append(items, item)
	}

	return items, nil
}

func shortRegistryName(full string) string {
	parts := strings.Split(full, "/")
	last := parts[len(parts)-1]
	last = strings.TrimPrefix(last, "server-")
	last = strings.TrimSuffix(last, "-mcp")
	last = strings.TrimSuffix(last, "-server")
	last = strings.ReplaceAll(last, "_", "-")
	return last
}

func registryCommand(registryType, identifier string) string {
	switch registryType {
	case "npm":
		return fmt.Sprintf("npx -y %s", identifier)
	case "pypi":
		return fmt.Sprintf("uvx %s", identifier)
	case "go":
		return fmt.Sprintf("go run %s", identifier)
	default:
		return fmt.Sprintf("npx -y %s", identifier)
	}
}
