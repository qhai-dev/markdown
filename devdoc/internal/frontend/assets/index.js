'use strict';

/* global default_theme, default_dark_theme, default_light_theme, hljs, ClipboardJS */

// Fix back button cache problem
window.onunload = function() { };

/**
 * Helper for global keypress handlers so they don't trigger when certain elements are active.
 * @returns {boolean} True if the keypress handler should be skipped.
 */
function mdbook_something_else_has_focus(e) {
    // Check composedPath in case the event happened from something generated
    // from the shadowDOM.
    const target = e.composedPath()[0] || e.target;
    // If this is the `checkbox-img` input which has the focus, we want to handle it here.
    if (target.classList.contains('checkbox-img')) {
        return false;
    }
    return /^(?:input|select|textarea)$/i.test(target.nodeName);
}

(function codeSnippets() {
    // Syntax highlighting Configuration
    hljs.configure({
        tabReplace: '    ', // 4 spaces
        languages: [], // Languages used for auto-detection
    });

    const code_nodes = Array
        .from(document.querySelectorAll('code'))
        // Don't highlight `inline code` blocks in headers.
        .filter(function(node) {
            return !node.parentElement.classList.contains('header');
        });

    code_nodes.forEach(function(block) {
        hljs.highlightBlock(block);
    });

    // Adding the hljs class gives code blocks the color css
    // even if highlighting doesn't apply
    code_nodes.forEach(function(block) {
        block.classList.add('hljs');
    });

    Array.from(document.querySelectorAll('code.hljs')).forEach(function(block) {

        const lines = Array.from(block.querySelectorAll('.boring'));
        // If no lines were hidden, return
        if (!lines.length) {
            return;
        }
        block.classList.add('hide-boring');

        const buttons = document.createElement('div');
        buttons.className = 'buttons';
        buttons.innerHTML = '<button title="Show hidden lines" \
aria-label="Show hidden lines"></button>';

        // add expand button
        const pre_block = block.parentNode;
        pre_block.insertBefore(buttons, pre_block.firstChild);

        buttons.firstChild.addEventListener('click', function(e) {
            if (this.title === 'Show hidden lines') {
                this.title = 'Hide lines';
                this.setAttribute('aria-label', e.target.title);

                block.classList.remove('hide-boring');
            } else if (this.title === 'Hide lines') {
                this.title = 'Show hidden lines';
                this.setAttribute('aria-label', e.target.title);

                block.classList.add('hide-boring');
            }
        });
    });

    Array.from(document.querySelectorAll('pre code')).forEach(function(block) {
        const pre_block = block.parentNode;
        let buttons = pre_block.querySelector('.buttons');
        if (!buttons) {
            buttons = document.createElement('div');
            buttons.className = 'buttons';
            pre_block.insertBefore(buttons, pre_block.firstChild);
        }

        const clipButton = document.createElement('button');
        clipButton.className = 'clip-button';
        clipButton.title = 'Copy to clipboard';
        clipButton.setAttribute('aria-label', clipButton.title);
        clipButton.innerHTML = '<i class="tooltiptext"></i>';

        buttons.insertBefore(clipButton, buttons.firstChild);
    });
})();

(function themes() {
    const html = document.querySelector('html');
    const themeToggleButton = document.getElementById('mdbook-theme-toggle');
    const themeColorMetaTag = document.querySelector('meta[name="theme-color"]');

    // Note: the per-theme stylesheet switching for the hljs themes and the
    // github-markdown variants was removed with the Writer front-end fork
    // (css/writer.css owns the syntax palette and both theme states). The
    // backup stylesheets are still emitted — see templates/index.html.

    function get_saved_theme() {
        let theme = null;
        try {
            theme = localStorage.getItem('mdbook-theme');
        } catch {
            // ignore error.
        }
        return theme;
    }

    function delete_saved_theme() {
        localStorage.removeItem('mdbook-theme');
    }

    function get_theme() {
        const theme = get_saved_theme();
        if (theme === null || theme === undefined) {
            if (typeof default_dark_theme === 'undefined') {
                // A customized index.hbs might not define this, so fall back to
                // old behavior of determining the default on page load.
                return default_theme;
            }
            return window.matchMedia('(prefers-color-scheme: dark)').matches
                ? default_dark_theme
                : default_light_theme;
        }
        return theme;
    }

    let previousTheme = default_theme;
    function set_theme(theme, store = true) {
        setTimeout(function() {
            themeColorMetaTag.content = getComputedStyle(document.documentElement).backgroundColor;
        }, 1);

        if (store) {
            try {
                localStorage.setItem('mdbook-theme', theme);
            } catch {
                // ignore error.
            }
        }

        html.classList.remove(previousTheme);
        html.classList.add(theme);
        previousTheme = theme;
    }

    const query = window.matchMedia('(prefers-color-scheme: dark)');
    query.onchange = function() {
        // set_theme(get_theme(), false);
    };

    // Set theme.
    // set_theme(get_theme(), false);
})();

(function chapterNavigation() {
    function zoomOutImages() {
        for (const elem of Array.from(document.querySelectorAll('input.checkbox-img'))) {
            elem.checked = false;
        }
    }

    document.addEventListener('keydown', function(e) {
        if (e.altKey ||
            e.ctrlKey ||
            e.metaKey ||
            window.search && window.search.hasFocus() ||
            mdbook_something_else_has_focus(e)
        ) {
            return;
        }

        const html = document.querySelector('html');

        function next() {
            const nextButton = document.querySelector('.nav-chapters.next');
            if (nextButton) {
                window.location.href = nextButton.href;
            }
        }
        function prev() {
            const previousButton = document.querySelector('.nav-chapters.previous');
            if (previousButton) {
                window.location.href = previousButton.href;
            }
        }
        function showHelp() {
            const container = document.getElementById('mdbook-help-container');
            const overlay = document.getElementById('mdbook-help-popup');
            container.style.display = 'flex';

            // Clicking outside the popup will dismiss it.
            const mouseHandler = event => {
                if (overlay.contains(event.target)) {
                    return;
                }
                if (event.button !== 0) {
                    return;
                }
                event.preventDefault();
                event.stopPropagation();
                document.removeEventListener('mousedown', mouseHandler);
                hideHelp();
            };

            // Pressing esc will dismiss the popup.
            const escapeKeyHandler = event => {
                if (event.key === 'Escape') {
                    event.preventDefault();
                    event.stopPropagation();
                    document.removeEventListener('keydown', escapeKeyHandler, true);
                    hideHelp();
                }
            };
            document.addEventListener('keydown', escapeKeyHandler, true);
            document.getElementById('mdbook-help-container')
                .addEventListener('mousedown', mouseHandler);
        }
        function hideHelp() {
            document.getElementById('mdbook-help-container').style.display = 'none';
        }

        // Usually needs the Shift key to be pressed
        switch (e.key) {
        case '?':
            e.preventDefault();
            showHelp();
            break;
        case 'Escape':
            zoomOutImages();
            break;
        }

        // Rest of the keys are only active when the Shift key is not pressed
        if (e.shiftKey) {
            return;
        }

        switch (e.key) {
        case 'ArrowRight':
            e.preventDefault();
            if (html.dir === 'rtl') {
                prev();
            } else {
                next();
            }
            break;
        case 'ArrowLeft':
            e.preventDefault();
            if (html.dir === 'rtl') {
                next();
            } else {
                prev();
            }
            break;
        }
    });
})();
