package web

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
)

const (
	distDir   = "dist"
	indexFile = "index.html"
)

var ErrFrontendNotBuilt = errors.New("frontend not built: run `task build:web`")

//go:embed all:dist
var dist embed.FS

func Assets() (fs.FS, error) {
	assets, err := fs.Sub(dist, distDir)
	if err != nil {
		return nil, fmt.Errorf("open embedded %s: %w", distDir, err)
	}
	if _, err := fs.Stat(assets, indexFile); err != nil {
		return assets, ErrFrontendNotBuilt
	}
	return assets, nil
}
