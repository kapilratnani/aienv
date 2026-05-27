package main

import (
	"embed"
	"io/fs"

	"github.com/kapilratnani/aienv/internal/assets"
)

//go:embed curated/mcps.yaml curated/skills.yaml
var curatedFS embed.FS

func init() {
	assets.Init(curatedFS)
}

func curatedDir() fs.FS {
	return curatedFS
}
