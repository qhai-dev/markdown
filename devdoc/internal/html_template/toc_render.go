package html_template

import (
	"path/filepath"
	"strings"

	"github.com/qhai-dev/devdoc/internal/model"
)

// renderTocSidebar computes the sidebar HTML for use via `{{.TocHTML}}`.
// It walks the chapter tree directly: container-only chapters (directory
// wrappers produced by the file-tree walker) render as `<li class="chapter-item expanded">`
// with no link, just their children; regular chapters emit `<a>` to
// the rendered HTML page. No section numbers / Part titles / Separators.
//
// isTocHTML controls the link `target=` attribute — true for the no-JS
// iframe fallback (kept here for parity), though iframe rendering has
// since been removed and the only call site passes false.
//
// Output mirrors the legacy `<ol class="chapter"><li>` shape so the
// existing CSS in chrome.css (`chapter-item`, `chapter-link-wrapper`,
// `chapter-fold-toggle`) keeps working unchanged.
func renderTocSidebar(chapters []*model.Chapter, foldEnable bool, foldLevel int, noSectionLabel bool, isTocHTML bool) string {
	var out strings.Builder
	out.WriteString(`<ol class="chapter">`)
	if len(chapters) > 0 {
		renderSidebarItems(&out, chapters, 1, foldEnable, foldLevel, noSectionLabel, isTocHTML)
	}
	out.WriteString("</li></ol>")
	return out.String()
}

// renderSidebarItems walks siblings, writing `<li>` wrappers + nested
// `<ol class="section">` groups. depth is the current nesting level
// (1 = top-level).
func renderSidebarItems(out *strings.Builder, items []*model.Chapter, depth int, foldEnable bool, foldLevel int, noSectionLabel bool, isTocHTML bool) {
	for _, ch := range items {
		if ch == nil {
			continue
		}
		expanded := !foldEnable || depth-1 < foldLevel
		writeSidebarLiOpen(out, expanded)

		if ch.IsContainer {
			// Container-only chapter: don't write a link; recurse directly into
			// its children inside an inner <ol class="section">.
			out.WriteString(`<span class="chapter-link-wrapper"><strong>`)
			out.WriteString(htmlEscape(ch.Name))
			out.WriteString(`</strong>`)
			if len(ch.SubItems) > 0 {
				out.WriteString(`<ol class="section">`)
				renderSidebarItems(out, ch.SubItems, depth+1, foldEnable, foldLevel, noSectionLabel, isTocHTML)
				out.WriteString("</li></ol>")
			}
			out.WriteString(`</span>`)
			out.WriteString("</li>")
			continue
		}

		out.WriteString(`<span class="chapter-link-wrapper">`)
		if !ch.IsDraft() && ch.Path != "" {
			out.WriteString(`<a href="`)
			href := filepath.ToSlash(ch.Path)
			href = strings.TrimPrefix(href, "./")
			out.WriteString(htmlEscape(withHTMLExtension(href)))
			if isTocHTML {
				out.WriteString(`" target="_parent">`)
			} else {
				out.WriteString(`">`)
			}
		} else {
			out.WriteString("<span>")
		}

		if !noSectionLabel {
			// Section labels removed in 2026-08-16. Histograms would go here.
			_ = noSectionLabel
		}
		out.WriteString(htmlEscape(ch.Name))
		if !ch.IsDraft() && ch.Path != "" {
			out.WriteString("</a>")
		} else {
			out.WriteString("</span>")
		}

		hasSub := len(ch.SubChapters()) > 0
		if hasSub && foldEnable {
			out.WriteString(`<a class="chapter-fold-toggle"><div>❱</div></a>`)
		}
		out.WriteString("</span>")

		if len(ch.SubItems) > 0 {
			out.WriteString(`<ol class="section">`)
			renderSidebarItems(out, ch.SubItems, depth+1, foldEnable, foldLevel, noSectionLabel, isTocHTML)
			out.WriteString("</li></ol>")
		}
		_ = hasSub
	}
}

func writeSidebarLiOpen(out *strings.Builder, expanded bool) {
	out.WriteString(`<li class="chapter-item `)
	if expanded {
		out.WriteString("expanded ")
	}
	out.WriteString(`">`)
}

// withHTMLExtension replaces a path's extension with `.html`, matching
// Rust's Path::with_extension.
func withHTMLExtension(path string) string {
	slash := strings.LastIndexByte(path, '/')
	dot := strings.LastIndexByte(path, '.')
	if dot > slash {
		return path[:dot] + ".html"
	}
	return path + ".html"
}

func htmlEscape(s string) string {
	r := strings.NewReplacer(
		`&`, `&amp;`,
		`<`, `&lt;`,
		`>`, `&gt;`,
		`"`, `&#34;`,
		`'`, `&#39;`,
	)
	return r.Replace(s)
}
