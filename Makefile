GIT_VERSION ?= $(shell git describe --tags --always --dirty="-dev")
DATE ?= $(shell date -u '+%Y-%m-%d %H:%M UTC')
VERSION_FLAGS := -X "main.version=$(GIT_VERSION)" -X "main.BuildTime=$(DATE)"

build:
	go build -ldflags='$(VERSION_FLAGS)' ./...

lint:
	go vet ./...
	golangci-lint run -E revive,gofmt ./...

test:
	go test -race -v -coverprofile="c.out" ./...
	go tool cover -func="c.out"

docs:
	cd documentation && mkdocs build
