// Package plugin implements mdBook's preprocessor and renderer extension
// protocol. The JSON shapes in this file mirror what the Rust mdBook
// implementation serialises over stdin/stdout (see
// crates/mdbook-preprocessor/src/lib.rs and crates/mdbook-renderer/src/lib.rs).
// Keeping this wire format separate from the internal book model lets the
// latter evolve without breaking plugin compatibility.
//
// Note (2026-08-16): the wire shape was simplified alongside the
// switch from `[chapters]` + section-numbering to front-matter + filesystem
// walk. External plugin shipping is currently frozen (see cmd.go header).
// A future plugin integration round will need to consume the new shape.
package plugin

import (
	"mdbook-go/internal/model"
)

// WireBook is the externally visible form of a book. It is serialised as
// the second element of the preprocessor input tuple and as the standalone
// preprocessor output.
//
// Items form a tree of WireChapter; container-only chapters (used purely
// for sidebar nesting) are tagged with IsContainer and have an empty
// Content plus no Path.
type WireBook struct {
	Items []WireChapter `json:"items"`
}

// WireChapter mirrors the internal Chapter (subset of fields that plugins
// need).
type WireChapter struct {
	Name        string        `json:"name"`
	Content     string        `json:"content"`
	Path        string        `json:"path,omitempty"`
	SourcePath  string        `json:"source_path,omitempty"`
	IsContainer bool          `json:"is_container,omitempty"`
	SubItems    []WireChapter `json:"sub_items,omitempty"`
}

// WireConfig mirrors crates/mdbook-core/src/config.rs::Config. Field tags are
// snake_case to match serde's defaults; the dynamic `output` and
// `preprocessor` maps stay as raw JSON so nested tables remain opaque to
// plugins.
type WireConfig struct {
	Package      PackageConfig  `json:"book"`
	Build        BuildConfig    `json:"build"`
	Output       map[string]any `json:"output"`
	Preprocessor map[string]any `json:"preprocessor"`
}

// PackageConfig mirrors mdbook-core's BookConfig.
type PackageConfig struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	Language      string `json:"language"`
	TextDirection string `json:"text-direction"`
	Root          string `json:"src"`
}

// BuildConfig mirrors mdbook-core's BuildConfig.
type BuildConfig struct {
	BuildDir                string   `json:"build-dir"`
	ExtraWatchDirs          []string `json:"extra-watch-dirs"`
	CreateMissing           bool     `json:"create-missing"`
	UseDefaultPreprocessors bool     `json:"use-default-preprocessors"`
}

// WirePreprocessorContext is the JSON handed to an external preprocessor on
// stdin. See crates/mdbook-preprocessor/src/lib.rs::PreprocessorContext.
type WirePreprocessorContext struct {
	Root          string     `json:"root"`
	Config        WireConfig `json:"config"`
	Renderer      string     `json:"renderer"`
	MdbookVersion string     `json:"mdbook_version"`
	// chapter_titles is internal to mdBook and is skipped on the wire (the
	// Rust side marks it with #[serde(skip)]).
	ChapterTitles map[string]string `json:"-"`
}

// WireRenderContext is the JSON handed to an external renderer. See
// crates/mdbook-renderer/src/lib.rs::RenderContext.
type WireRenderContext struct {
	Version     string     `json:"version"`
	Root        string     `json:"root"`
	Book        WireBook   `json:"book"`
	Config      WireConfig `json:"config"`
	Destination string     `json:"destination"`
	// chapter_titles is internal and skipped.
	ChapterTitles map[string]string `json:"-"`
}

// MdbookVersion is the version string embedded in both context types. The
// Rust side reads CARGO_PKG_VERSION; the Go side stamps whatever value is
// passed to Build.
const MdbookVersion = "0.1.0-m3"

// ToWireBook converts the internal book model into the wire representation.
func ToWireBook(b *model.Book) WireBook {
	if b == nil {
		return WireBook{}
	}
	out := WireBook{Items: make([]WireChapter, 0, len(b.Items))}
	for _, it := range b.Items {
		out.Items = append(out.Items, toWireChapter(it))
	}
	return out
}

func toWireChapter(c *model.Chapter) WireChapter {
	wc := WireChapter{
		Name:        c.Name,
		Content:     c.Content,
		IsContainer: c.IsContainer,
		SubItems:    make([]WireChapter, 0, len(c.SubItems)),
	}
	if !c.IsContainer && !c.IsDraft() {
		wc.Path = c.Path
		wc.SourcePath = c.SourcePath
	}
	for _, sub := range c.SubItems {
		wc.SubItems = append(wc.SubItems, toWireChapter(sub))
	}
	return wc
}

// FromWireBook turns a wire representation back into the internal model. It
// is used after an external preprocessor has returned its modified book.
func FromWireBook(w WireBook) *model.Book {
	out := model.NewBook()
	for _, it := range w.Items {
		out.Items = append(out.Items, fromWireChapter(it))
	}
	return out
}

func fromWireChapter(wc WireChapter) *model.Chapter {
	if wc.IsContainer {
		children := make([]*model.Chapter, 0, len(wc.SubItems))
		for _, sub := range wc.SubItems {
			children = append(children, fromWireChapter(sub))
		}
		return model.NewContainerChapter(wc.Name, wc.SourcePath, children)
	}
	ch := model.NewChapter(wc.Name, wc.Path)
	ch.Content = wc.Content
	for _, sub := range wc.SubItems {
		ch.SubItems = append(ch.SubItems, fromWireChapter(sub))
	}
	return ch
}

// ToWireConfig converts the internal config to the wire shape.
func ToWireConfig(c *model.Config) WireConfig {
	if c == nil {
		return WireConfig{}
	}
	return WireConfig{
		Package: PackageConfig{
			Title:         c.Package.Title,
			Description:   c.Package.Description,
			Language:      c.Package.Language,
			TextDirection: c.Package.TextDirection,
			Root:          c.Package.Root,
		},
		Build: BuildConfig{
			BuildDir:                c.Build.BuildDir,
			ExtraWatchDirs:          c.Build.ExtraWatchDirs,
			CreateMissing:           c.Build.CreateMissing,
			UseDefaultPreprocessors: c.Build.UseDefaultPreprocessors,
		},
		Output:       c.Output,
		Preprocessor: c.Preprocessor,
	}
}
