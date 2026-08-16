#!/usr/bin/env bash
# Smoke-test the devdoc_server target: start the server, curl the rendered
# intro page, kill it. Exit 0 on success.
#
# Invoked by bazel as a sh_test with runfiles symlinked under $TEST_SRCDIR.
# $1 is the rootpath-relative binary location (e.g. tests/basic/book_server);
# we resolve it via $TEST_SRCDIR.

set -euo pipefail

BOOK_SERVER_REL="$1"
# bazel puts runfiles under either $TEST_SRCDIR/ or $TEST_SRCDIR/_main/
# depending on workspace setup. Probe for both.
for prefix in "$TEST_SRCDIR" "$TEST_SRCDIR/_main"; do
    if [ -e "$prefix/$BOOK_SERVER_REL" ]; then
        BOOK_SERVER="$prefix/$BOOK_SERVER_REL"
        BOOK_DIR_BASE="$prefix"
        break
    fi
done
if [ -z "${BOOK_SERVER:-}" ]; then
    echo "FAIL: book_server not found under $TEST_SRCDIR/$BOOK_SERVER_REL" >&2
    exit 1
fi
PORT="${PORT:-34567}"  # picked to avoid common dev-server ports

# Resolve the book directory alongside the server binary.
BOOK_DIR="$BOOK_DIR_BASE/$(dirname "$BOOK_SERVER_REL")"
if [ ! -f "$BOOK_DIR/devdoc.yaml" ]; then
    echo "FAIL: book dir not found at $BOOK_DIR" >&2
    exit 1
fi
cd "$BOOK_DIR"

# Start server in the background. devdoc serve flags must come AFTER the
# subcommand (cobra parses them as serve subcommand flags).
"$BOOK_SERVER" serve --hostname=127.0.0.1 --port="$PORT" --dir=. &
SERVER_PID=$!

# Wait for the server to come up (up to ~30s for a fresh build).
ATTEMPTS=30
URL="http://127.0.0.1:$PORT/intro.html"
until curl -sf "$URL" >/dev/null 2>&1; do
    ATTEMPTS=$((ATTEMPTS - 1))
    if [ "$ATTEMPTS" -le 0 ]; then
        echo "FAIL: devdoc server did not become ready within 30s" >&2
        kill "$SERVER_PID" 2>/dev/null || true
        exit 1
    fi
    sleep 1
done

# Fetch and validate the rendered intro page.
BODY=$(curl -sf "$URL")
if ! echo "$BODY" | grep -qi "Welcome"; then
    echo "FAIL: intro page missing expected content" >&2
    echo "$BODY" | head -20 >&2
    kill "$SERVER_PID" 2>/dev/null || true
    exit 1
fi

# Also verify the index page works.
INDEX_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:$PORT/")
if [ "$INDEX_STATUS" != "200" ]; then
    echo "FAIL: index page returned $INDEX_STATUS" >&2
    kill "$SERVER_PID" 2>/dev/null || true
    exit 1
fi

kill "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true

echo "OK: devdoc_server served intro.html with expected content"
exit 0