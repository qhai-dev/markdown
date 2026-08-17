package runner

import (
	"bytes"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// FrontMatter holds the YAML metadata parsed from the leading `---\n...\n---`
// block of a chapter file. Title is required (the build errors out if
// missing) — Index is optional (nil means "use lex sort").
//
// `Draft` is also optional and only used for documentation; the build
// currently treats every encountered .md file as a chapter.
type FrontMatter struct {
	Title string `yaml:"title"`
	Index *int   `yaml:"index"`
	Draft bool   `yaml:"draft"`
}

// frontMatterDelim matches a single `---` line at the start of the file.
// We accept both the leading `---\n` opener and the matching `---\n`
// closer; everything in between is YAML.
var frontMatterDelim = []byte("---\n")

// parseFrontMatter inspects content; if it begins with a `---\n` YAML
// block, it returns the parsed struct plus the remainder of the file
// (after stripping the closing `---\n`). If the file has no front-matter
// at all it returns (zero, content, nil). A malformed block returns an
// error so the build does not silently ignore a typo.
func parseFrontMatter(content []byte) (FrontMatter, []byte, error) {
	var fm FrontMatter
	if !bytes.HasPrefix(content, frontMatterDelim) {
		return fm, content, nil
	}
	rest := content[len(frontMatterDelim):]
	closeIdx := bytes.Index(rest, frontMatterDelim)
	if closeIdx < 0 {
		return fm, content, errors.New("front-matter: missing closing `---`")
	}
	raw := rest[:closeIdx]
	body := rest[closeIdx+len(frontMatterDelim):]
	if err := yaml.Unmarshal(raw, &fm); err != nil {
		return fm, content, fmt.Errorf("front-matter: %w", err)
	}
	// Strip BOM if the body has one — goldmark itself handles BOM, but
	// doing it here keeps our "what gets stored as Content" consistent.
	body = bytes.TrimPrefix(body, []byte{0xEF, 0xBB, 0xBF})
	return fm, body, nil
}
