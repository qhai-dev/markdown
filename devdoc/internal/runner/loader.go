// Package runner drives the book build pipeline. Loader builds an in-memory
// book tree from a source directory using front-matter metadata on each
// Markdown file. See walk.go for the directory-traversal rules.
package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/qhai-dev/devdoc/internal/model"
)

// MDBook is the central handle for a loaded book.
type MDBook struct {
	Root   string
	Config *model.Config
	Book   *model.Book
}

// Load resolves the book root, reads devdoc.yaml, and walks the source
// directory to build the chapter tree. The walker is the authoritative
// source of structure now; devdoc.yaml only carries package / build /
// output / preprocessor configuration.
func Load(root string) (*MDBook, error) {
	if root == "" {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	cfg, err := model.LoadBook(abs)
	if err != nil {
		return nil, err
	}
	cfg.SetRoot(abs)

	book := model.NewBook()
	if cfg.Package.Root != "" {
		items, err := walkSourceDir(cfg.Package.Root)
		if err != nil {
			return nil, fmt.Errorf("walk source: %w", err)
		}
		book.Items = items
	}

	return &MDBook{Root: abs, Config: cfg, Book: book}, nil
}

// BuildDir returns the absolute output directory.
func (m *MDBook) BuildDir() string {
	bd := m.Config.Build.BuildDir
	if !filepath.IsAbs(bd) {
		bd = filepath.Join(m.Root, bd)
	}
	return bd
}

// SourceDir returns the absolute source directory.
func (m *MDBook) SourceDir() string { return m.Config.Package.Root }

// PathToRoot returns the relative path from a chapter's source location
// back to the book root.
func (m *MDBook) PathToRoot() string {
	srcRel, err := filepath.Rel(m.Root, m.Config.Package.Root)
	if err != nil {
		return ""
	}
	depth := strings.Count(filepath.Clean(srcRel), string(os.PathSeparator))
	if depth == 0 {
		return "./"
	}
	return strings.Repeat("../", depth)
}
