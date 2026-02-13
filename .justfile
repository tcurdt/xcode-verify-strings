# justfile

set quiet

NAME := "xcode-verify-strings"
VERSION := "v0.0.2"

DEFAULT_TARGETS := "darwin-arm64" # darwin-amd64 linux-arm64"

default:
    just --list

# build targets (default: all)
build *targets:
    #!/usr/bin/env bash
    set -euo pipefail
    for t in {{ if targets == "" { DEFAULT_TARGETS } else { targets } }}; do
        OS="${t%-*}"
        ARCH="${t#*-}"
        echo "building bin: [$OS $ARCH]"
        mkdir -p "bin/$t"
        GOOS="$OS" GOARCH="$ARCH" go build -o "bin/$t/{{NAME}}"
        touch "bin/$t/{{NAME}}"
    done

# package targets (default: all)
package *targets: (build targets)
    #!/usr/bin/env bash
    set -euo pipefail
    for t in {{ if targets == "" { DEFAULT_TARGETS } else { targets } }}; do
        OS="${t%-*}"
        ARCH="${t#*-}"
        echo "building dist: [$OS $ARCH]"
        mkdir -p dist
        tar -C "bin/$t" \
            -czf "dist/{{NAME}}-{{VERSION}}-$t.tgz" \
            .
        touch "dist/{{NAME}}-{{VERSION}}-$t.tgz"
    done

clean:
    rm -rf dist bin

test: (build)
    go test

run:
    go run $(find . -name "*.go" -and -not -name "*_test.go")

install: (build "darwin-arm64")
    mkdir -p "$HOME/.bin"
    cp "bin/darwin-arm64/{{NAME}}" "$HOME/.bin/{{NAME}}"

release: (package)
    git tag -a {{VERSION}} -m "releasing {{VERSION}}"
    git push --tags origin master
