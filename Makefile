PHONY: build snapshot lint test docs
GIT_VERSION ?= $(shell git describe --tags --always --dirty="-dev")
DATE ?= $(shell date -u '+%Y-%m-%d %H:%M UTC')
VERSION_FLAGS := -s -w -X "main.version=$(GIT_VERSION)" -X "main.date=$(DATE)"
.DEFAULT_GOAL := build
DOCS_DIR := $(CURDIR)/documentation

build:
	go build -trimpath -ldflags='$(VERSION_FLAGS) -extldflags -static' ./cmd/...

snapshot:
	cd ./cmd/cloudflare-utils; goreleaser --snapshot --clean --skip=publish,sign

lint:
	golangci-lint run --config .golangci-lint.yml ./cmd/...

test:
	echo "No tests yet"
	exit 0
	go test -race -coverprofile="coverage.out" ./...
	go tool cover -func="c.out"

docs:
	cd documentation && mkdocs build

docs-serve:
	cd documentation && docker run --rm -v $(DOCS_DIR):/docs -p 8000:8000 ghcr.io/squidfunk/mkdocs-material:9.5.4