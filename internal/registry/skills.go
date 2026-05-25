package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type SkillsDotSh struct {
	client *http.Client
}

type skillsSearchResponse struct {
	Query    string `json:"query"`
	Skills   []skillsResult `json:"skills"`
	Count    int    `json:"count"`
}

type skillsResult struct {
	ID       string `json:"id"`
	SkillID  string `json:"skillId"`
	Name     string `json:"name"`
	Installs int    `json:"installs"`
}

func NewSkillsDotSh() *SkillsDotSh {
	return &SkillsDotSh{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SkillsDotSh) Search(ctx context.Context, query string, limit int) ([]SkillItem, error) {
	u, _ := url.Parse("https://skills.sh/api/search")
	q := u.Query()
	q.Set("q", query)
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying skills.sh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("skills.sh returned status %d", resp.StatusCode)
	}

	var searchResp skillsSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	items := make([]SkillItem, 0, len(searchResp.Skills))
	for _, sk := range searchResp.Skills {
		items = append(items, SkillItem{
			Name:        sk.Name,
			Package:     extractPackage(sk.ID),
			Description: sk.Name,
			Installs:    sk.Installs,
		})
	}

	return items, nil
}

func extractPackage(id string) string {
	parts := splitN(id, "/", 3)
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return id
}

func splitN(s, sep string, n int) []string {
	out := make([]string, 0, n)
	start := 0
	for i := 0; i < n-1 && start < len(s); i++ {
		if idx := indexOf(s[start:], sep); idx >= 0 {
			out = append(out, s[:start+idx])
			start += idx + len(sep)
		}
	}
	out = append(out, s[start:])
	return out
}

func indexOf(s, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
