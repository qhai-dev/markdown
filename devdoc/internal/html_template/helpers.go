package html_template

import (
	"html/template"
	"path/filepath"
	"strings"

	"github.com/qhai-dev/devdoc/internal/model"
)

// Env carries the per-build state that the template helpers need.
type Env struct {
	// Resources maps a logical asset name to its hashed filename.
	Resources map[string]string
	// Path is the current page path inside the book.
	Path string
	// Chapters is the full chapter tree (with IsContainer flags) walked by
	// TocHTML into the chapter list sidebar.
	Chapters []*model.Chapter
	// FoldEnable / FoldLevel / NoSectionLabel mirror the corresponding
	// [output.html.fold] settings; NoSectionLabel is a no-op since
	// section labels were dropped in 2026-08-16, kept for config compat.
	FoldEnable     bool
	FoldLevel      int
	NoSectionLabel bool
	// Content is the rendered chapter HTML.
	Content template.HTML
	// LiveReloadEndpoint is the URL fragment the live-reload WebSocket
	// connects to.
	LiveReloadEndpoint template.URL
	// FragmentMap is the redirect-fragment map for this page.
	FragmentMap template.JS
}

// Resource implements `{{resource "name"}}`.
func (e *Env) Resource(name string) string {
	resolved, ok := e.Resources[name]
	if !ok {
		resolved = name
	}
	return pathToRoot(e.Path) + resolved
}

// TocHTML returns the chapter list sidebar HTML for use inside a template
// via `{{.TocHTML}}`. The iframe fallback was removed; this is the only
// consumer path.
func (e *Env) TocHTML() template.HTML {
	return template.HTML(renderTocSidebar(e.Chapters, e.FoldEnable, e.FoldLevel, e.NoSectionLabel, false))
}

// pathToRoot returns the relative prefix needed to reach the output root
// from the given page path.
func pathToRoot(p string) string {
	if p == "" {
		return ""
	}
	dir := filepath.ToSlash(filepath.Dir(p))
	if dir == "." || dir == "" {
		return ""
	}
	depth := strings.Count(dir, "/") + 1
	return strings.Repeat("../", depth)
}
