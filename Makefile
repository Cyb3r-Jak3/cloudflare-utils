PHONY: build snapshot lint test docs
GIT_VERSION ?= $(shell git describe --tags --always --dirty="-dev")
DATE ?= $(shell date -u '+%Y-%m-%d %H:%M UTC')
VERSION_FLAGS := -s -w -X "main.version=$(GIT_VERSION)" -X "main.date=$(DATE)"
.DEFAULT_GOAL := build
DOCS_DIR := $(CURDIR)/documentation
GOTESTSUM_JUNITFILE ?= $(CURDIR)/-junit.xml
build:
	go build -trimpath -ldflags='$(VERSION_FLAGS) -extldflags -static' ./cmd/...

snapshot:
	goreleaser --snapshot --clean --skip=publish,sign

lint:
	golangci-lint run --config .golangci-lint.yml ./cmd/...

test:
	@gotestsum --format testname --junitfile junit.xml -- -coverprofile=cover.out ./...

docs:
	cd documentation && mkdocs build

docs-serve:
	cd documentation && docker run --rm -v $(DOCS_DIR):/docs -p 8000:8000 ghcr.io/squidfunk/mkdocs-material:9.5.4