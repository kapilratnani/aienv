package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
		items = append(items, MCPServerItem{
			Name:        s.Server.Name,
			Description: s.Server.Description,
			Type:        "local",
			Command:     fmt.Sprintf("npx -y @modelcontextprotocol/server-%s", s.Server.Name),
		})
	}

	return items, nil
}
