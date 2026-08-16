// Package model defines the in-memory book model and the typed/dynamic
// configuration objects that drive a build. The shapes mirror
// crates/mdbook-core/src/book.rs so that JSON serialisation stays
// compatible with the existing preprocessor and renderer protocol.
package model

import (
	"path/filepath"
	"strings"
)

// Chapter is one node in the book tree.
//
// The tree is now discovered by walking the source directory; there are no
// `[chapters]` lists, no PartTitle / Separator variants, and no section
// numbering. A Chapter either renders as a page (`IsContainer == false`)
// or acts purely as a directory container in the sidebar
// (`IsContainer == true`, empty Content, no Path).
type Chapter struct {
	Name        string     `json:"name"`
	Content     string     `json:"content"`
	Path        string     `json:"path"`
	SourcePath  string     `json:"source_path"`
	IsContainer bool       `json:"is_container,omitempty"`
	SubItems    []*Chapter `json:"sub_items"`
}

// NewChapter constructs a regular (non-container) Chapter with the given
// name and path. The source path is derived from the path when not
// explicitly set.
func NewChapter(name, path string) *Chapter {
	return &Chapter{
		Name:       name,
		Path:       path,
		SourcePath: path,
	}
}

// NewDraftChapter creates a placeholder chapter that is not backed by a
// file. The renderer will skip it during build but it can still appear in
// the sidebar (typically via its parent listing it).
func NewDraftChapter(name string) *Chapter {
	return &Chapter{Name: name, Path: ""}
}

// NewContainerChapter wraps a directory into a non-rendering node used
// purely for sidebar nesting. Its `Name` is the directory basename and
// `SubItems` are the children discovered inside.
func NewContainerChapter(name string, subdirRel string, children []*Chapter) *Chapter {
	return &Chapter{
		Name:        name,
		IsContainer: true,
		SourcePath:  subdirRel,
		SubItems:    children,
	}
}

// SubChapters returns only the child items that are real (non-draft,
// non-container) chapters.
func (c *Chapter) SubChapters() []*Chapter {
	var out []*Chapter
	for _, item := range c.SubItems {
		if item == nil || item.IsContainer || item.IsDraft() {
			continue
		}
		out = append(out, item)
	}
	return out
}

// IsDraft reports whether the chapter has no backing path.
func (c *Chapter) IsDraft() bool {
	return c.Path == "" && !c.IsContainer
}

// HTMLPath returns the output file for a chapter, relative to the book
// root. The source path's directory structure is preserved and only the
// extension is replaced, so `guide/advanced/deep.md` becomes
// `guide/advanced/deep.html`.
func (c *Chapter) HTMLPath() string {
	if c.IsContainer || c.IsDraft() {
		return ""
	}
	base := strings.TrimSuffix(c.Path, filepath.Ext(c.Path))
	if base == "" {
		return ""
	}
	return base + ".html"
}

// Book is the ordered list of top-level items. Book.Items may contain a
// mix of regular chapters and directory-container chapters; both use the
// same *Chapter type with IsContainer flag for differentiation.
type Book struct {
	Items []*Chapter `json:"sections"`
}

// NewBook returns an empty Book.
func NewBook() *Book {
	return &Book{Items: []*Chapter{}}
}

// Chapters returns all renderable chapters in depth-first order.
// Container nodes and drafts are skipped.
func (b *Book) Chapters() []*Chapter {
	var out []*Chapter
	for _, item := range b.Items {
		out = append(out, collectChapters(item)...)
	}
	return out
}

func collectChapters(c *Chapter) []*Chapter {
	if c == nil || c.IsDraft() {
		return nil
	}
	if c.IsContainer {
		// Skip container itself but recurse into children.
		var out []*Chapter
		for _, ch := range c.SubItems {
			out = append(out, collectChapters(ch)...)
		}
		return out
	}
	out := []*Chapter{c}
	for _, ch := range c.SubItems {
		out = append(out, collectChapters(ch)...)
	}
	return out
}

// Iter visits the book in depth-first order. Containers are recursed into
// but not visited themselves; drafts are also skipped. The callback may
// return false to stop iteration.
func (b *Book) Iter(fn func(*Chapter) bool) {
	for _, item := range b.Items {
		if !walkChapter(item, fn) {
			return
		}
	}
}

func walkChapter(c *Chapter, fn func(*Chapter) bool) bool {
	if c == nil {
		return true
	}
	if !c.IsContainer && !c.IsDraft() {
		if !fn(c) {
			return false
		}
	}
	for _, ch := range c.SubItems {
		if !walkChapter(ch, fn) {
			return false
		}
	}
	return true
}
