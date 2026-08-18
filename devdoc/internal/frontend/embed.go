// Package frontend embeds the default mdBook front-end assets and the
// production Go template (index.html).
//
// The assets/ files (CSS + JS) are a verbatim copy of
// crates/mdbook-html/front-end/ in the Rust workspace; index.html is Go-only
// (the html/template source that replaced the Rust theme templates). They
// live here because Go's //go:embed directive cannot reach outside the
// module, and the Go port must be able to produce byte-identical output
// without the Rust tree being present. Refresh the front-end with:
//
//	rm -rf mdbook-go/internal/frontend && mkdir -p mdbook-go/internal/frontend \
//	  && cp -R crates/mdbook-html/front-end/. mdbook-go/internal/frontend/
//
// (then restore this file — index.html is Go-only with no Rust counterpart).
package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:assets index.html
var files embed.FS

// FS exposes the embedded front-end tree.
func FS() fs.FS {
	return files
}

// MustRead returns the contents of an embedded file, panicking when the file is
// missing. Every call site uses a path that is present at compile time.
func MustRead(name string) []byte {
	data, err := files.ReadFile(name)
	if err != nil {
		panic("assets: " + err.Error())
	}
	return data
}
