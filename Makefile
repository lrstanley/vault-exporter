.DEFAULT_GOAL := build

GOPATH := $(shell go env | grep GOPATH | sed 's/GOPATH="\(.*\)"/\1/')
PATH := $(GOPATH)/bin:$(PATH)
export $(PATH)

# enable Go 1.11.x module support.
export GO111MODULE=on

BINARY=vault-exporter
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null | sed -r "s:^v::g")

help:
		@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

snapshot: clean fetch ## Generate a snapshot release.
		$(GOPATH)/bin/goreleaser release --snapshot --skip-validate --skip-publish

release: clean fetch ## Generate a release, but don't publish to GitHub.
		$(GOPATH)/bin/goreleaser release --skip-validate --skip-publish

publish: clean fetch ## Generate a release, and publish to GitHub.
		$(GOPATH)/bin/goreleaser release

fetch: ## Fetches the necessary dependencies to build.
		test -f $(GOPATH)/bin/goreleaser || go get -u -v github.com/goreleaser/goreleaser
		go mod download
		go mod tidy
		go mod vendor

clean: ## Cleans up generated files/folders from the build.
		/bin/rm -rfv "dist/" "${BINARY}"

build: fetch clean ## Compile and generate a binary.
		go build -ldflags '-d -s -w' -tags netgo -installsuffix netgo -v -x -o "${BINARY}"

install-tools: install-go install-go-releaser
.PHONY: install-tools

install-go:
	wget -nv -P /tmp https://dl.google.com/go/go$(GO_VERSION).$(OS)-$(ARCH).tar.gz
	tar -C ~/ -xzf /tmp/go$(GO_VERSION).$(OS)-$(ARCH).tar.gz
	rm -r /tmp/go$(GO_VERSION).$(OS)-$(ARCH).tar.gz
.PHONY: install-go

install-goreleaser:
	## Stupid non-standard release format... Linux not linux and x86_64 not amd64.
	wget -nv -P /tmp/ https://github.com/goreleaser/goreleaser/releases/download/v$(GORELEASER_VERSION)/goreleaser_Linux_x86_64.tar.gz
	tar -C ~/bin -xzf /tmp/goreleaser_Linux_x86_64.tar.gz goreleaser
	rm -r /tmp/goreleaser_Linux_x86_64.tar.gz
.PHONY: install-goreleaser
