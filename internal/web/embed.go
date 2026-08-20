package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

// StaticFS returns the embedded static file system rooted at static/.
func StaticFS() (fs.FS, error) {
	return fs.Sub(staticFS, "static")
}

// MustStaticFS returns static FS or panics.
func MustStaticFS() fs.FS {
	f, err := StaticFS()
	if err != nil {
		panic(err)
	}
	return f
}

// FileServer returns an http.Handler for embedded assets.
func FileServer() http.Handler {
	sub := MustStaticFS()
	return http.FileServer(http.FS(sub))
}
