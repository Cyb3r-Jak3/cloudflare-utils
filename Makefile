PHONY: build snapshot lint test docs
GIT_VERSION ?= $(shell git describe --tags --always --dirty="-dev")
DATE ?= $(shell date -u '+%Y-%m-%d %H:%M UTC')
VERSION_FLAGS := -s -w -X "main.version=$(GIT_VERSION)" -X "main.date=$(DATE)"
.DEFAULT_GOAL := build

build:
	go build -trimpath -ldflags='$(VERSION_FLAGS) -extldflags -static' ./cmd/...

snapshot:
	cd ./cmd/cloudflare-utils; goreleaser --snapshot  --clean --skip=publish,sign

lint:
	golangci-lint run --config .golangci-lint.yml ./...

test:
	echo "No tests yet"
	exit 0
	go test -race -v -coverprofile="c.out" ./...
	go tool cover -func="c.out"

docs:
	cd documentation && mkdocs build
