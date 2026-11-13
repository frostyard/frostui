.PHONY: verify verify-static verify-template check ci fmt lint lint-version-check test test-wasm race cross snapshot tidy corrections-base bump

GOCMD ?= go

# The kit builds without CGO everywhere: MyGo loads GTK, OpenGL and the
# platform's libraries at run time with purego.
CGO_ENABLED = 0

# Pinned golangci-lint release, read from mise.toml, the single source of
# every tool pin: `mise install` provisions it locally and CI installs it with
# jdx/mise-action, verified against mise.lock. Compared with `golangci-lint
# version --short`, which prints the bare "MAJOR.MINOR.PATCH".
GOLANGCI_LINT_VERSION := $(strip $(shell sed -n 's/^golangci-lint = "\(.*\)"/\1/p' mise.toml))
# The Go release from go.mod's go line, the only Go pin. golangci-lint must be
# built with a Go of at least this language version (MAJOR.MINOR).
GO_VERSION := $(strip $(shell sed -n 's/^go \(.*\)/\1/p' go.mod))
GO_LANGUAGE := $(shell echo "$(GO_VERSION)" | cut -d. -f1,2)
# The major version of the Node on PATH (mise.toml pins it): the template gate
# and the js/wasm tests run on it.
NODE_MAJOR := $(shell node -p "process.versions.node.split('.')[0]" 2>/dev/null)

# The GOOS/GOARCH pairs the kit builds for. internal/repocheck's
# TestKitCrossCompiles and the cross-build matrix of
# .github/workflows/ci.yml list the same pairs
# (TestCrossTargetsAreListedEverywhere).
CROSS_TARGETS ?= linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64 js/wasm

# Where make snapshot writes the examples' PNGs.
SNAPSHOT_DIR ?= build/snapshots

# The tests need no display: MyGo's Tester draws on the CPU with Pango and
# cairo. Unsetting DISPLAY and WAYLAND_DISPLAY proves it on a desktop too.
HEADLESS = env -u DISPLAY -u WAYLAND_DISPLAY

fmt:
	gofmt -s -w .

tidy:
	$(GOCMD) mod tidy

test:
	$(HEADLESS) CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) test ./...

# The tests again as js/wasm under Node: the only off-Linux platform this
# host can run, so the _other.go fallbacks run too.
test-wasm:
	@if ! command -v node >/dev/null 2>&1; then echo "FAIL: test-wasm requires Node.js (run: mise install)"; exit 1; fi
	GOOS=js GOARCH=wasm CGO_ENABLED=0 $(GOCMD) test -exec="$$($(GOCMD) env GOROOT)/lib/wasm/go_js_wasm_exec" ./...

# The race detector needs CGO and a C compiler.
race:
	$(HEADLESS) CGO_ENABLED=1 $(GOCMD) test -race -short ./...

# Build and vet every package, tests included, for each target.
cross:
	@for target in $(CROSS_TARGETS); do \
		goos=$${target%/*}; goarch=$${target#*/}; \
		echo "==> cross: $$goos/$$goarch"; \
		GOOS=$$goos GOARCH=$$goarch CGO_ENABLED=0 $(GOCMD) build ./... || exit 1; \
		GOOS=$$goos GOARCH=$$goarch CGO_ENABLED=0 $(GOCMD) vet ./... || exit 1; \
	done

# Write the examples' frames, light and dark, as PNGs. Copy
# $(SNAPSHOT_DIR)/commandcenter.png, threepane.png and sourcelist.png to
# docs/images/ to refresh the docs' pictures.
snapshot:
	@mkdir -p $(SNAPSHOT_DIR)
	$(HEADLESS) FROSTUI_SNAPSHOT_DIR=$(abspath $(SNAPSHOT_DIR)) CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) test -count=1 -run 'Snapshot$$' -v ./examples/...

# Lint (fails when the mise.toml pin is missing or the installed release
# differs), as the host and as each other system the kit builds for, so the
# platform files (_linux.go, _other.go) are linted too
lint:
	$(MAKE) --no-print-directory lint-version-check
	CGO_ENABLED=$(CGO_ENABLED) golangci-lint run
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) golangci-lint run
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) golangci-lint run
	GOOS=js GOARCH=wasm CGO_ENABLED=$(CGO_ENABLED) golangci-lint run

# Fail unless the installed golangci-lint is the mise.toml pin and was built
# with a Go no older than go.mod's go line
lint-version-check:
	@test -n "$(GOLANGCI_LINT_VERSION)" || { echo "mise.toml pins no golangci-lint"; exit 1; }
	@installed="$$(golangci-lint version --short 2>/dev/null)" || { \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) is required (not installed; run: mise install)"; exit 1; }; \
	if [ "$$installed" != "$(GOLANGCI_LINT_VERSION)" ]; then \
		echo "expected golangci-lint $(GOLANGCI_LINT_VERSION), found $$installed (run: mise install)"; \
		exit 1; \
	fi; \
	built="$$(golangci-lint version 2>/dev/null | sed -n 's/.*built with go\([0-9]*\.[0-9]*\).*/\1/p')"; \
	if [ -n "$$built" ] && [ "$$(printf '%s\n%s\n' "$(GO_LANGUAGE)" "$$built" | sort -V | head -1)" != "$(GO_LANGUAGE)" ]; then \
		echo "golangci-lint $(GOLANGCI_LINT_VERSION) was built with go$$built, older than go.mod's go$(GO_LANGUAGE): bump golangci-lint first"; \
		exit 1; \
	fi

# Check the agent surfaces, docs index, risk paths, corrections and vendored
# copies (needs Node.js 22 or later, from mise.toml)
verify-template:
	@if ! command -v node >/dev/null 2>&1; then echo "FAIL: make verify requires Node.js 22 or later and make; run: mise install"; exit 1; fi
	@if [ "$(NODE_MAJOR)" = "" ] || [ "$(NODE_MAJOR)" -lt 22 ]; then echo "FAIL: make verify requires Node.js 22 or later and make (found: $$(node --version 2>/dev/null || echo missing)); run: mise install"; exit 1; fi
	node scripts/check-template.mjs
	node --test test/*.test.mjs

# The non-mutating static checks: tidy diff, vet, gofmt, the pinned lint and
# the template gate
verify-static:
	@echo "==> verify: go.mod is tidy"
	$(GOCMD) mod tidy -diff
	@echo "==> verify: go vet"
	CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) vet ./...
	@echo "==> verify: gofmt"
	test -z "$$(gofmt -s -l .)"
	@echo "==> verify: lint"
	$(MAKE) --no-print-directory lint
	@echo "==> verify: template, docs index, risk paths, corrections, vendored copies"
	$(MAKE) --no-print-directory verify-template

# The credential-free gate that leaves git status unchanged: verify-static,
# then every test (the import, API and cross-build guards included)
verify: verify-static
	@echo "==> verify: tests"
	$(MAKE) --no-print-directory test

# Developer gate: format, then verify
check: fmt verify

# What CI's verify job runs, in fail-fast order: verify, the corrections
# check against BASE when it is set (make ci BASE=origin/main), the js/wasm
# tests, the race detector and the cross-build of every target.
ci: verify
	$(MAKE) --no-print-directory corrections-base
	@echo "==> ci: js/wasm tests"
	$(MAKE) --no-print-directory test-wasm
	@echo "==> ci: race detector"
	$(MAKE) --no-print-directory race
	@echo "==> ci: cross-build"
	$(MAKE) --no-print-directory cross
	@echo "==> ci passed"

# The records in .memory/corrections/ against BASE: no record removed or
# renamed, only an empty promoted_to filled, and the old single corrections
# file does not come back. A BASE that does not resolve to a commit is an
# error, never a silent pass.
corrections-base:
	@if [ -z "$(BASE)" ]; then exit 0; fi; \
	if ! git rev-parse --verify --quiet "$(BASE)^{commit}" >/dev/null; then \
		echo "FAIL: BASE $(BASE) does not resolve to a commit (fetch it first)"; \
		exit 1; \
	fi; \
	node scripts/check-corrections.mjs --base "$(BASE)"

# Tag the next version (svu, .svu.yml) and push the tag. The maintainer runs
# it; agents never tag or push a tag.
bump:
	$(MAKE) --no-print-directory verify
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working directory is not clean. Please commit or stash changes before bumping version."; \
		exit 1; \
	fi
	@version=$$(svu next); \
		git tag -a $$version -m "Version $$version"; \
		echo "Tagged version $$version"; \
		echo "Pushing tag $$version to origin..."; \
		git push origin $$version
