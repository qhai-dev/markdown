package runner

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"mdbook-go/internal/model"
)

// walkSourceDir scans srcDir and returns the book tree. Each directory
// contributes a non-rendering "container" Chapter whose SubItems are the
// children discovered inside, sorted by front-matter `index:` ascending
// (numeric) or by basename (lex) when no index is set. Items with
// `index:` come BEFORE items without it so authors can mix freely.
//
// Files that are not `.md` or `.MD` are ignored. Hidden directories
// (starting with `.`) are skipped entirely.
//
// Returns nil if srcDir is empty or missing (the caller decides whether
// that's an error).
func walkSourceDir(srcDir string) ([]*model.Chapter, error) {
	if _, err := os.Stat(srcDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	children, err := walkDir(srcDir, "")
	if err != nil {
		return nil, err
	}
	return children, nil
}

// walkDir reads srcDir and returns a sorted slice of Chapters — one for
// each .md file, plus one container-Chapter for each non-empty
// subdirectory. dirRel is the path relative to the source root, used
// only as the SourcePath on container nodes.
func walkDir(srcDir, dirRel string) ([]*model.Chapter, error) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", srcDir, err)
	}
	type entry struct {
		chapter  *model.Chapter
		hasIndex bool
		indexN   int
		sortKey  string
	}
	var out []entry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(srcDir, name)
		childRel := filepath.Join(dirRel, name)
		if e.IsDir() {
			grandkids, err := walkDir(full, childRel)
			if err != nil {
				return nil, err
			}
			if len(grandkids) == 0 {
				continue // skip empty directories
			}
			out = append(out, entry{
				chapter:  model.NewContainerChapter(name, childRel, grandkids),
				sortKey:  name,
				hasIndex: false,
			})
			continue
		}
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		data, err := os.ReadFile(full)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", full, err)
		}
		fm, body, err := parseFrontMatter(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", childRel, err)
		}
		// Files without front-matter (no leading `---`) are skipped — they
		// are conventionally include-source fragments ({{#include}} targets,
		// not chapters).
		if !bytes.HasPrefix(data, []byte("---\n")) {
			continue
		}
		if strings.TrimSpace(fm.Title) == "" {
			return nil, fmt.Errorf("%s: front-matter `title:` is required", childRel)
		}
		ch := model.NewChapter(fm.Title, childRel)
		ch.Content = string(body)
		hasIdx := fm.Index != nil
		var idxN int
		if hasIdx {
			idxN = *fm.Index
		}
		out = append(out, entry{
			chapter:  ch,
			hasIndex: hasIdx,
			indexN:   idxN,
			sortKey:  name,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].hasIndex != out[j].hasIndex {
			return out[i].hasIndex // indexed items first
		}
		if out[i].hasIndex && out[j].hasIndex {
			if out[i].indexN != out[j].indexN {
				return out[i].indexN < out[j].indexN
			}
		}
		return out[i].sortKey < out[j].sortKey
	})
	result := make([]*model.Chapter, 0, len(out))
	for _, e := range out {
		result = append(result, e.chapter)
	}
	return result, nil
}
