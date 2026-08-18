// The `devdoc watch` subcommand and the filesystem-watching engine it
// shares with `serve`: rebuild the book whenever a source file changes.
// Both the command and the poll-based engine live in this one file,
// mirroring mdbook's bin: `src/cmd/watch/poller.rs`. The fsnotify-based
// native backend (native.rs) is not ported.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/qhai-dev/devdoc/internal/model"
	"github.com/qhai-dev/devdoc/internal/runner"

	"github.com/spf13/cobra"
)

// newWatchCommand implements the `devdoc watch` subcommand.
func newWatchCommand() *cobra.Command {
	var dir, dest string

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "rebuild on file changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(dir, dest)
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "book root")
	cmd.Flags().StringVar(&dest, "dest-dir", "", "output directory (overrides devdoc.yaml build-dir)")

	return cmd
}

// run is the entry point for `devdoc watch`. It installs a SIGINT
// handler so Ctrl-C exits cleanly and stops the poll loop.
func runWatch(dir, dest string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	return Watch(ctx, dir, WatchOptions{
		UpdateConfig: func(m *runner.MDBook) {
			if dest != "" {
				m.Config.Build.BuildDir = dest
			}
		},
	})
}

// WatchOptions configures a Watch() invocation. UpdateConfig is called on
// every reload so callers can re-apply CLI flags (e.g. --dest-dir). It
// runs after the book has been reloaded from disk, so any config the
// user wrote to devdoc.yaml in the meantime is visible.
//
// PostBuild is invoked after a successful rebuild; serve uses it to
// broadcast a reload message over its WebSocket.
type WatchOptions struct {
	UpdateConfig func(*runner.MDBook)
	PostBuild    func()
}

// Watch blocks until ctx is cancelled, rebuilding the book every time a
// watched file changes. The first build runs synchronously before
// entering the wait loop so users see a populated output directory
// before the watcher starts idling.
//
// On any rebuild error the loop continues — the watcher should not die
// because of a transient devdoc.yaml syntax error or a stale chapter. The
// error is logged to stderr and the next change retries the build.
func Watch(ctx context.Context, dir string, opts WatchOptions) error {
	if opts.UpdateConfig == nil {
		opts.UpdateConfig = func(*runner.MDBook) {}
	}
	if opts.PostBuild == nil {
		opts.PostBuild = func() {}
	}

	m, err := runner.Load(dir)
	if err != nil {
		return fmt.Errorf("watch: load: %w", err)
	}
	opts.UpdateConfig(m)
	if err := m.Build(); err != nil {
		// Don't bail — the first build may legitimately fail (e.g. the
		// user is in the middle of editing devdoc.yaml). The next change
		// will retry.
		fmt.Fprintf(os.Stderr, "watch: initial build failed: %v\n", err)
	} else {
		opts.PostBuild()
	}

	roots := CollectWatchRoots(m)
	return watchPoll(ctx, m, roots, opts)
}

func watchPoll(ctx context.Context, m *runner.MDBook, roots []WatchRoot, opts WatchOptions) error {
	pw := NewPollWatcher(m.Root)
	pw.SetRoots(roots)
	// Prime the cache so the first tick doesn't flag every file as new.
	_ = pw.Scan()
	ticker := time.NewTicker(PollTick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			changed := pw.Scan()
			if len(changed) == 0 {
				continue
			}
			if !rebuild(m.Root, opts) {
				continue
			}
			opts.PostBuild()
		}
	}
}

// rebuild reloads the book from disk, re-applies the CLI-level config
// override, and runs a fresh build. The boolean return reports whether
// the build succeeded so the caller knows whether to call PostBuild.
func rebuild(dir string, opts WatchOptions) bool {
	m, err := runner.Load(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "watch: reload failed: %v\n", err)
		return false
	}
	opts.UpdateConfig(m)
	if err := m.Build(); err != nil {
		fmt.Fprintf(os.Stderr, "watch: build failed: %v\n", err)
		return false
	}
	return true
}

// =====================================================================
// PollWatcher: poll-based filesystem watcher. Direct port of
// src/cmd/watch/poller.rs::Watcher. The fsnotify-based native backend
// (native.rs) is not ported — this implementation works on any
// filesystem, including network mounts and Docker bind mounts where the
// kernel does not reliably deliver change events.
//
// PollWatcher is safe for concurrent Scan() calls but the typical use is
// a single goroutine that scans on a fixed timer.
// =====================================================================

// PollWatcher is the watcher state. Polls one or more roots and reports
// changed paths back to the caller.
type PollWatcher struct {
	roots []WatchRoot
	prev  map[string]pathStat
	mu    sync.Mutex
}

// WatchRoot describes one path the watcher polls on each tick. The
// `Extensions` filter restricts which files inside the path trigger a
// rebuild; nil means "accept every regular file". This is what lets
// `CollectWatchRoots` declare `docs/` to only react to `.md` changes
// while still honouring theme / css / devdoc.yaml / extra-watch-dirs
// edits — those roots use a nil filter and accept everything.
type WatchRoot struct {
	Path       string
	Extensions map[string]bool // nil = no filter; ".md" only is `{".md": true}`
}

// pathStat is the cache key for a single file. Directories are skipped
// during scan (their mtime is unreliable on most filesystems) so only
// regular file entries live in the cache.
type pathStat struct {
	mtime time.Time
	size  int64
}

// mtime returns the file's modification time as reported by os.FileInfo.
// This mirrors src/cmd/watch/poller.rs::Watcher::scan, which uses
// `meta.modified().unwrap_or(SystemTime::UNIX_EPOCH)` directly with no
// platform-specific branch.
//
// Earlier revisions of this file tried to fall back to
// syscall.Stat_t.Ctim when ModTime() returned zero — a paranoid measure
// that did not actually compile on macOS (the field is named Ctimespec
// on Darwin, Ctim on Linux) and would have produced spurious rebuilds
// anyway, since ctime tracks metadata changes (chmod / rename / owner)
// rather than content. Reverting to the simpler "ModTime or zero"
// semantic keeps the Go watcher bit-for-bit equivalent to the Rust one.
func mtime(info os.FileInfo) time.Time {
	return info.ModTime()
}

// NewPollWatcher creates a watcher rooted at the given book directory.
// All paths under SetRoots() are scanned; there is no .gitignore filtering
// — the previous Gitignore matcher was removed in 2026-08-16 because it
// covered only a partial subset of gitignore syntax.
func NewPollWatcher(bookRoot string) *PollWatcher {
	return &PollWatcher{
		prev: map[string]pathStat{},
	}
}

// SetRoots configures the paths the watcher will poll on each tick.
// This is called from Watch() once the book has been loaded so the
// watcher's roots stay in sync with the source dir, theme dir,
// additional-css/js, devdoc.yaml, and [build] extra-watch-dirs.
//
// Each root carries an optional Extensions filter so, for example, the
// source dir can be limited to .md changes only while other roots
// accept every file.
func (w *PollWatcher) SetRoots(roots []WatchRoot) {
	w.mu.Lock()
	defer w.mu.Unlock()
	cleaned := make([]WatchRoot, 0, len(roots))
	for _, r := range roots {
		if r.Path == "" {
			continue
		}
		abs, err := filepath.Abs(r.Path)
		if err != nil {
			continue
		}
		cleaned = append(cleaned, WatchRoot{Path: abs, Extensions: r.Extensions})
	}
	w.roots = cleaned
}

// Scan walks the configured roots and returns the set of paths that
// changed since the previous scan. Both new paths and removed paths are
// reported; the caller is expected to treat every entry as "rebuild
// needed" without inspecting the diff.
func (w *PollWatcher) Scan() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	next := make(map[string]pathStat)
	for _, root := range w.roots {
		if _, err := os.Stat(root.Path); err != nil {
			continue
		}
		w.walkInto(root, next)
	}

	var changed []string
	for path, stat := range next {
		if old, ok := w.prev[path]; !ok || old != stat {
			changed = append(changed, path)
		}
	}
	for path := range w.prev {
		if _, ok := next[path]; !ok {
			changed = append(changed, path)
		}
	}
	w.prev = next
	sort.Strings(changed)
	return changed
}

// walkInto records regular-file stats inside root, applying root's
// Extensions filter. A nil filter means "accept every regular file";
// a non-nil filter restricts to entries whose file extension appears
// in the filter map (e.g. {".md": true} only accepts Markdown).
//
// Symlinks are followed (matching Rust WalkDir::follow_links(true)).
func (w *PollWatcher) walkInto(root WatchRoot, into map[string]pathStat) {
	_ = filepath.Walk(root.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// A single unreadable entry should not abort the whole
			// scan; skip and continue. This matches the Rust
			// `filter_map` that drops entries whose metadata call
			// fails with a debug log.
			return nil
		}
		// Resolve symlinks so the cache key is stable across watches.
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil
			}
			fi, err := os.Stat(resolved)
			if err != nil {
				return nil
			}
			info = fi
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if !root.accepts(path) {
			return nil
		}
		into[path] = pathStat{mtime: mtime(info), size: info.Size()}
		return nil
	})
}

// accepts reports whether root would record a change to path under
// its Extensions filter. A nil filter accepts everything; otherwise the
// file's extension (lowercased, with leading dot) must be a key in the
// map.
func (r WatchRoot) accepts(path string) bool {
	if r.Extensions == nil {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return false // plain filenames are not accepted when a filter is set
	}
	return r.Extensions[ext]
}

// Tick is the default poll interval. The Rust implementation uses 1s;
// matching that keeps the user-facing experience consistent across
// implementations.
const PollTick = time.Second

// CollectWatchRoots computes the set of paths PollWatcher should scan.
// It mirrors src/cmd/watch/poller.rs::Watcher::set_roots in spirit, but
// has one important refinement: the source dir is restricted to `.md`
// changes only — random non-md files in `docs/` (images, scratch notes,
// stray binaries) no longer trigger rebuilds. Theme / config /
// additional-css / additional-js / extra-watch-dirs are all left with
// a nil filter and accept every file change.
//
// HTML() can fail (e.g. malformed [output.html] table); on error we
// fall back to an empty additional-css/js list and let the render path
// surface the error later. The watcher itself doesn't need the config
// to be perfectly valid — it just needs enough to know what to scan.
func CollectWatchRoots(m *runner.MDBook) []WatchRoot {
	mdOnly := map[string]bool{".md": true}
	roots := []WatchRoot{
		{Path: m.SourceDir(), Extensions: mdOnly},
		{Path: filepath.Join(m.Root, model.ConfigFileName)},
	}
	if htmlCfg, err := m.Config.HTML(); err == nil {
		roots = append(roots, WatchRoot{Path: htmlCfg.ThemeDir(m.Root)})
		for _, css := range htmlCfg.AdditionalCSS {
			roots = append(roots, WatchRoot{Path: filepath.Join(m.Root, css)})
		}
		for _, js := range htmlCfg.AdditionalJS {
			roots = append(roots, WatchRoot{Path: filepath.Join(m.Root, js)})
		}
	}
	for _, dir := range m.Config.Build.ExtraWatchDirs {
		roots = append(roots, WatchRoot{Path: filepath.Join(m.Root, dir)})
	}
	return roots
}
