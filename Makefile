GIT_VERSION ?= $(shell git describe --tags --always --dirty="-dev")
DATE ?= $(shell date -u '+%Y-%m-%d %H:%M UTC')
VERSION_FLAGS := -X "main.version=$(GIT_VERSION)" -X "main.BuildTime=$(DATE)"

build:
	go build -ldflags='$(VERSION_FLAGS)' ./...

lint:
	go vet ./...
	golanglint-ci run -E revive,gofmt ./...