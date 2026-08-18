package html_template

import (
	"html/template"
	"path/filepath"
	"strings"

	"github.com/qhai-dev/devdoc/internal/model"
	"github.com/qhai-dev/devdoc/pkg/fs"
)

// RenderData is the typed template data the production .html files consume.
// It embeds Env so the {{.Resource}} helper is available as a promoted method.
//
// Optional fields (LiveReloadEndpoint, etc.) are left at their zero value
// when the corresponding legacy data-map key is absent; the .html files
// gate them with `{{if .X}}` so an empty value renders as "feature off".
type RenderData struct {
	*Env

	// Identity / chrome.
	Language           string
	Title              string
	ChapterTitle       string
	BookTitle          string
	Description        string
	DefaultTheme       string
	PreferredDarkTheme string
	TextDirection      string
	PathToRoot         string

	// NavMode selects the keyboard navigation script ("vim" or "normal").
	NavMode string

	// Page flags.
	IsIndex bool
	BaseURL string

	// Optional features.
	SearchEnabled  bool
	SearchJS       bool
	MathJaxSupport bool

	// Resources.
	AdditionalCSS []string
	AdditionalJS  []string

	// Live reload (template.URL so html/template treats it as URL-safe).
	LiveReloadEndpoint template.URL

	// Fragment map (template.JS so JS string escapes correctly).
	FragmentMap template.JS

	// Chapter navigation.
	Previous *Nav
	Next     *Nav

	// Content is the rendered chapter HTML body. The .html template renders
	// it inside <main> via {{.Content}} (the legacy data map calls it
	// "content"; this field is the typed RenderData view).
	Content template.HTML

	// Git integration.
	GitRepositoryURL       string
	GitRepositoryEditURL   string
	GitRepositoryIconClass string
	GitRepositoryIcon      string
}

// Nav is the shape of the Previous / Next links on a page.
type Nav struct {
	Title string
	Link  string
}

// makeData builds the shared template context in the loose
// map[string]any form. Callers that want a typed RenderData use BuildRenderData.
func makeData(ctx *Context, cfg *model.HtmlConfig, th *Theme) map[string]any {
	data := map[string]any{}

	data["language"] = ctx.Config.Package.Language
	if data["language"] == "" {
		data["language"] = "en"
	}
	data["text_direction"] = ctx.Config.Package.RealizedTextDirection()
	data["book_title"] = ctx.Config.Package.Title
	data["description"] = ctx.Config.Package.Description
	if cfg.LiveReloadEndpoint != "" {
		data["live_reload_endpoint"] = cfg.LiveReloadEndpoint
	}
	data["default_theme"] = cfg.DefaultThemeName()
	data["preferred_dark_theme"] = cfg.PreferredDarkThemeName()
	if cfg.MathJaxSupport {
		data["mathjax_support"] = true
	}
	if len(cfg.AdditionalCSS) > 0 {
		data["additional_css"] = relativeToRoot(ctx.Root, cfg.AdditionalCSS)
	}
	if len(cfg.AdditionalJS) > 0 {
		data["additional_js"] = relativeToRoot(ctx.Root, cfg.AdditionalJS)
	}
	data["nav_mode"] = cfg.Mode

	search := cfg.EffectiveSearch()
	data["search_enabled"] = search.Enable
	data["search_js"] = search.Enable && search.CopyJS

	if cfg.GitRepositoryURL != "" {
		data["git_repository_url"] = cfg.GitRepositoryURL
	}
	data["git_repository_icon"] = cfg.GitRepositoryIconName()
	data["git_repository_icon_class"] = cfg.GitRepositoryIconClass()

	data["fold_enable"] = cfg.Fold.Enable
	data["fold_level"] = int(cfg.Fold.Level)
	data["chapters"] = ctx.Book.Items
	return data
}

// BuildRenderData converts the loose data map into the typed RenderData
// struct that the production .html files consume. The struct embeds Env so
// {{.Resource "x"}} works as a promoted method.
func BuildRenderData(data map[string]any) RenderData {
	out := RenderData{
		Language:               asString(data, "language"),
		Title:                  asString(data, "title"),
		ChapterTitle:           asString(data, "chapter_title"),
		BookTitle:              asString(data, "book_title"),
		Description:            asString(data, "description"),
		DefaultTheme:           asString(data, "default_theme"),
		PreferredDarkTheme:     asString(data, "preferred_dark_theme"),
		TextDirection:          asString(data, "text_direction"),
		PathToRoot:             asString(data, "path_to_root"),
		NavMode:                asString(data, "nav_mode"),
		BaseURL:                asString(data, "base_url"),
		SearchEnabled:          asBool(data, "search_enabled"),
		SearchJS:               asBool(data, "search_js"),
		MathJaxSupport:         asBool(data, "mathjax_support"),
		IsIndex:                asBool(data, "is_index"),
		GitRepositoryURL:       asString(data, "git_repository_url"),
		GitRepositoryEditURL:   asString(data, "git_repository_edit_url"),
		GitRepositoryIconClass: asString(data, "git_repository_icon_class"),
		GitRepositoryIcon:      asString(data, "git_repository_icon"),
	}
	out.Env = &Env{}
	out.Env.Path = asString(data, "path")
	out.Env.FoldEnable = asBool(data, "fold_enable")
	out.Env.FoldLevel = asInt(data, "fold_level")
	out.Env.NoSectionLabel = asBool(data, "no_section_label")
	if v, ok := data["live_reload_endpoint"]; ok {
		out.LiveReloadEndpoint = template.URL(asStringFromAny(v))
	}
	if v, ok := data["fragment_map"]; ok {
		out.FragmentMap = template.JS(asStringFromAny(v))
	}
	if v, ok := data["content"]; ok {
		if h, ok := v.(template.HTML); ok {
			out.Content = h
		} else {
			out.Content = template.HTML(asStringFromAny(v))
		}
	}
	if v, ok := data["previous"]; ok {
		if m, ok := v.(map[string]any); ok {
			out.Previous = &Nav{
				Title: asString(m, "title"),
				Link:  asString(m, "link"),
			}
		}
	}
	if v, ok := data["next"]; ok {
		if m, ok := v.(map[string]any); ok {
			out.Next = &Nav{
				Title: asString(m, "title"),
				Link:  asString(m, "link"),
			}
		}
	}
	if v, ok := data["additional_css"]; ok {
		out.AdditionalCSS = asStringSlice(v)
	}
	if v, ok := data["additional_js"]; ok {
		out.AdditionalJS = asStringSlice(v)
	}
	if v, ok := data["resources"]; ok {
		if m, ok := v.(map[string]string); ok {
			out.Env.Resources = m
		}
	}
	if v, ok := data["chapters"]; ok {
		if s, ok := v.([]*model.Chapter); ok {
			out.Env.Chapters = s
		}
	}
	return out
}

// asString / asBool / asInt / asStringSlice / asStringFromAny are tiny
// coercion helpers that tolerate the loose data maps.
func asString(data map[string]any, key string) string {
	if data == nil {
		return ""
	}
	v, ok := data[key]
	if !ok {
		return ""
	}
	return asStringFromAny(v)
}

func asStringFromAny(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func asBool(data map[string]any, key string) bool {
	if data == nil {
		return false
	}
	v, ok := data[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

func asInt(data map[string]any, key string) int {
	if data == nil {
		return 0
	}
	v, ok := data[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		if n == "" {
			return 0
		}
		var out int
		for _, c := range n {
			if c < '0' || c > '9' {
				return 0
			}
			out = out*10 + int(c-'0')
		}
		return out
	}
	return 0
}

func asStringSlice(v any) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, 0, len(s))
		for _, e := range s {
			if str, ok := e.(string); ok {
				out = append(out, str)
			}
		}
		return out
	}
	return nil
}

// relativeToRoot strips the book root prefix from user-specified asset paths.
func relativeToRoot(root string, paths []string) []any {
	out := make([]any, 0, len(paths))
	for _, p := range paths {
		if rel, err := filepath.Rel(root, p); err == nil && !strings.HasPrefix(rel, "..") {
			out = append(out, fs.ToURLPath(rel))
			continue
		}
		out = append(out, fs.ToURLPath(p))
	}
	return out
}
