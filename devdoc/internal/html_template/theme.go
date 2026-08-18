// This file resolves the asset set used to render a book: the embedded
// defaults, overridden file-by-file from the user's theme directory. It is a
// port of crates/mdbook-html/src/theme/mod.rs.
package html_template

import (
	"os"
	"path/filepath"

	static "github.com/qhai-dev/devdoc/internal/static"
)

// Theme carries the resolved contents of every themeable asset.
type Theme struct {
	ChromeCSS    []byte
	GeneralCSS   []byte
	VariablesCSS []byte

	JS               []byte
	HighlightCSS     []byte
	TomorrowNightCSS []byte
	AyuHighlightCSS  []byte
	HighlightJS      []byte
	ClipboardJS      []byte

	// NavVimJS / NavNormalJS are the keyboard navigation variants selected by
	// [output.html.mode]: nav-vim.js adds h/l chapter navigation.
	NavVimJS    []byte
	NavNormalJS []byte

	// OutlineRailJS drives the right-hand heading rail. It has no Rust
	// counterpart (see test-html-css/index.html for the reference demo).
	OutlineRailJS []byte
}

// Bundled front-end static.
var (
	// SearcherJS, MarkJS and ElasticlunrJS back the search UI.
	SearcherJS    = static.MustRead("searcher/searcher.js")
	MarkJS        = static.MustRead("searcher/mark.min.js")
	ElasticlunrJS = static.MustRead("searcher/elasticlunr.min.js")
)

// Default returns the embedded theme with no user overrides applied.
func Default() *Theme {
	return &Theme{
		ChromeCSS:              static.MustRead("css/chrome.css"),
		GeneralCSS:             static.MustRead("css/general.css"),
		VariablesCSS:           static.MustRead("css/variables.css"),
		JS:                     static.MustRead("js/book.js"),
		HighlightCSS:           static.MustRead("css/highlight.css"),
		TomorrowNightCSS:       static.MustRead("css/tomorrow-night.css"),
		AyuHighlightCSS:        static.MustRead("css/ayu-highlight.css"),
		HighlightJS:            static.MustRead("js/highlight.min.js"),
		ClipboardJS:            static.MustRead("js/clipboard.min.js"),
		NavVimJS:               static.MustRead("js/nav-vim.js"),
		NavNormalJS:            static.MustRead("js/nav-normal.js"),
		OutlineRailJS:          static.MustRead("js/outline-rail.js"),
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
		{"book.js", &t.JS},
		{"css/chrome.css", &t.ChromeCSS},
		{"css/general.css", &t.GeneralCSS},
		{"css/variables.css", &t.VariablesCSS},
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
