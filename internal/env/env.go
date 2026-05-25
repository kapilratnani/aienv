package env

type Env struct {
	Name        string              `yaml:"name"`
	Agent       string              `yaml:"agent"`
	Model       string              `yaml:"model,omitempty"`
	Description string              `yaml:"description,omitempty"`
	MCPServers  map[string]MCPServer `yaml:"mcp"`
	Skills      []Skill             `yaml:"skills"`
	Rules       []Rule              `yaml:"rules"`
}

type MCPServer struct {
	Type    string            `yaml:"type"`
	Command []string          `yaml:"command,omitempty"`
	URL     string            `yaml:"url,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type Skill struct {
	Name    string `yaml:"name"`
	Source  string `yaml:"source"`
	Package string `yaml:"package,omitempty"`
	Path    string `yaml:"path,omitempty"`
}

type Rule struct {
	Path string `yaml:"path"`
}
