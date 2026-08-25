BINARY := repo-knowledge
GITHUB_ADAPTER := repo-knowledge-github-adapter.sh
VERSION := $(shell tr -d '\n' < VERSION)
DIST := dist
GO ?= go

.PHONY: build test vet check eval-list release

build:
	CGO_ENABLED=0 $(GO) build -trimpath -o $(BINARY) ./cmd/repo-knowledge

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: test vet
	@test -z "$$($(GO)fmt -l .)" || { $(GO)fmt -l .; exit 1; }

eval-list:
	$(GO) run ./cmd/repo-knowledge-eval list

release:
	mkdir -p $(DIST)
	rm -f $(DIST)/$(BINARY)-* $(DIST)/LICENSE $(DIST)/SHA256SUMS
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags "-s -w" -o $(DIST)/$(BINARY)-linux-amd64 ./cmd/repo-knowledge
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags "-s -w" -o $(DIST)/$(BINARY)-linux-arm64 ./cmd/repo-knowledge
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags "-s -w" -o $(DIST)/$(BINARY)-darwin-amd64 ./cmd/repo-knowledge
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -trimpath -ldflags "-s -w" -o $(DIST)/$(BINARY)-darwin-arm64 ./cmd/repo-knowledge
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -trimpath -ldflags "-s -w" -o $(DIST)/$(BINARY)-windows-amd64.exe ./cmd/repo-knowledge
	install -m 0755 adapters/github/documentation-check.sh $(DIST)/$(GITHUB_ADAPTER)
	cp LICENSE $(DIST)/LICENSE
	cd $(DIST) && if command -v sha256sum >/dev/null 2>&1; then sha256sum $(BINARY)-* LICENSE > SHA256SUMS; else shasum -a 256 $(BINARY)-* LICENSE > SHA256SUMS; fi
	@printf 'built repository-knowledge %s release artifacts in %s\n' "$(VERSION)" "$(DIST)"
