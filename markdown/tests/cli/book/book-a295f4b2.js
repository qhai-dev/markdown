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
