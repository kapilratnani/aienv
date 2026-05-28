package docker

import (
	"fmt"

	"embed"
)

//go:embed *.Dockerfile
var dockerfiles embed.FS

func readDockerfile(agent string) ([]byte, error) {
	switch agent {
	case "opencode":
		return dockerfiles.ReadFile("opencode.Dockerfile")
	case "claude-code":
		return dockerfiles.ReadFile("claude.Dockerfile")
	default:
		return nil, fmt.Errorf("unknown agent %q for Dockerfile", agent)
	}
}
