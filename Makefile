PHONY: build snapshot lint test docs
GIT_VERSION ?= $(shell git describe --tags --always --dirty="-dev")
DATE ?= $(shell date -u '+%Y-%m-%d %H:%M UTC')
VERSION_FLAGS := -X "main.version=$(GIT_VERSION)" -X "main.date=$(DATE)"

build:
	go build -ldflags='$(VERSION_FLAGS)' ./cmd/...

snapshot:
	cd ./cmd/cloudflare-utils; goreleaser --snapshot --skip-publish --clean --skip-sign

lint:
	golangci-lint run --config .golangci-lint.yml ./...

test:
	go test -race -v -coverprofile="c.out" ./...
	go tool cover -func="c.out"

docs:
	cd documentation && mkdocs build
