package docker

import (
	"embed"
)

//go:embed sandbox.Dockerfile
var dockerfiles embed.FS

func BaseDockerfile() ([]byte, error) {
	return dockerfiles.ReadFile("sandbox.Dockerfile")
}
