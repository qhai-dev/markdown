// This file resolves the asset set used to render a book: the embedded
// defaults, overridden file-by-file from the user's theme directory. It is a
// port of crates/mdbook-html/src/theme/mod.rs.
package html_template

import (
	"os"
	"path/filepath"

	"github.com/qhai-dev/devdoc/internal/frontend"
)

// Theme carries the resolved contents of every themeable asset.
type Theme struct {
	ChromeCSS      []byte
	GeneralCSS     []byte
	VariablesCSS   []byte
	MarkdownCSS    []byte

	JS               []byte
	HighlightCSS     []byte
	TomorrowNightCSS []byte
	AyuHighlightCSS  []byte
	HighlightJS      []byte
	ClipboardJS      []byte

	// OutlineRailJS drives the right-hand heading rail. It has no Rust
	// counterpart (see test-html-css/index.html for the reference demo).
	OutlineRailJS []byte
}

// Bundled front-end static.
var (
	// SearcherJS, MarkJS and ElasticlunrJS back the search UI.
	SearcherJS    = frontend.MustRead("assets/searcher.js")
	MarkJS        = frontend.MustRead("assets/mark.min.js")
	ElasticlunrJS = frontend.MustRead("assets/elasticlunr.min.js")
)

// Default returns the embedded theme with no user overrides applied.
func Default() *Theme {
	return &Theme{
		ChromeCSS:              frontend.MustRead("assets/chrome.css"),
		GeneralCSS:             frontend.MustRead("assets/general.css"),
		VariablesCSS:           frontend.MustRead("assets/variables.css"),
		MarkdownCSS:            frontend.MustRead("assets/github-markdown.css"),
		JS:                     frontend.MustRead("assets/index.js"),
		HighlightCSS:           frontend.MustRead("assets/highlight.css"),
		TomorrowNightCSS:       frontend.MustRead("assets/tomorrow-night.css"),
		AyuHighlightCSS:        frontend.MustRead("assets/ayu-highlight.css"),
		HighlightJS:            frontend.MustRead("assets/highlight.min.js"),
		ClipboardJS:            frontend.MustRead("assets/clipboard.min.js"),
		OutlineRailJS:          frontend.MustRead("assets/outline-rail.js"),
	}
}

// New loads the defaults and then overrides individual files from themeDir.
// A missing directory or a missing file is not an error.
//
// As of 2026-08-08 we render via Go templates (this package), not hbs; the
// override table therefore no longer lists any *.hbs file. The bundled
// templates live in templates/ and have no user-extensible surface (yet).
// Overrides that remain: the JS bundle and the css/js assets, mirroring the
// chromium-side surface Rust mdBook still exposes.
func NewTheme(themeDir string) *Theme {
	t := Default()
	info, err := os.Stat(themeDir)
	if err != nil || !info.IsDir() {
		return t
	}

	overrides := []struct {
		rel  string
		dest *[]byte
	}{
		{"index.js", &t.JS},
		{"css/chrome.css", &t.ChromeCSS},
		{"css/general.css", &t.GeneralCSS},
		{"css/variables.css", &t.VariablesCSS},
		{"css/github-markdown.css", &t.MarkdownCSS},
		{"highlight.min.js", &t.HighlightJS},
		{"clipboard.min.js", &t.ClipboardJS},
		{"highlight.css", &t.HighlightCSS},
		{"tomorrow-night.css", &t.TomorrowNightCSS},
		{"ayu-highlight.css", &t.AyuHighlightCSS},
		{"outline-rail.js", &t.OutlineRailJS},
	}
	for _, o := range overrides {
		loadInto(filepath.Join(themeDir, filepath.FromSlash(o.rel)), o.dest)
	}

	return t
}

// loadInto replaces *dest with the file contents and reports whether it did.
func loadInto(path string, dest *[]byte) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	*dest = data
	return true
}
