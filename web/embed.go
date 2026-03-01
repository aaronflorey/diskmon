package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/* dist/assets/*
var distFS embed.FS

func Assets() fs.FS {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil
	}
	return sub
}
