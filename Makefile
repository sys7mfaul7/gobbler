# gobbler build targets. Run `make help`.
# Toolchains: see scripts/setup-*.sh

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -s -w -X main.version=$(VERSION)
DIST      := dist
# First-class targets; darwin/freebsd are best-effort.
PLATFORMS ?= linux/amd64 linux/arm64 windows/amd64 windows/arm64 android/arm64
BEST_EFFORT ?= darwin/amd64 darwin/arm64 freebsd/amd64
CMDS      := gobbler playlist-gobbler-aimp
JELLYFIN_DIR ?= ../playlist-gobbler.jellyfin
NEXTCLOUD_DIR := playlist-gobbler.nextcloud
ANDROID_API ?= 24

.PHONY: help all test vet fmt build cross cross-best-effort mobile-aar \
        nextcloud nextcloud-test jellyfin clean

help: ## list targets
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-20s %s\n",$$1,$$2}'

all: vet test cross nextcloud-test ## core checks + cross builds + nextcloud tests

fmt: ## gofmt check
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

vet: fmt ## go vet
	go vet ./...

test: ## go unit tests
	go test ./...

build: ## build for this machine into dist/
	@mkdir -p $(DIST)
	for c in $(CMDS); do CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST)/$$c ./cmd/$$c; done

cross: ## static binaries for first-class platforms
	@$(MAKE) --no-print-directory _cross P="$(PLATFORMS)"

cross-best-effort: ## darwin/freebsd builds
	@$(MAKE) --no-print-directory _cross P="$(BEST_EFFORT)"

_cross:
	@mkdir -p $(DIST)
	@set -e; for p in $(P); do os=$${p%/*}; arch=$${p#*/}; ext=; [ $$os = windows ] && ext=.exe; \
	  for c in $(CMDS); do echo "build $$c $$os/$$arch"; \
	    CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -ldflags "$(LDFLAGS)" \
	      -o $(DIST)/$$c-$$os-$$arch$$ext ./cmd/$$c; done; done

mobile-aar: ## Android AAR via gomobile (needs ANDROID_HOME + NDK)
	@mkdir -p $(DIST)
	gomobile bind -target android/arm64 -androidapi $(ANDROID_API) -o $(DIST)/gobbler-mobile.aar ./mobile

nextcloud: ## build the Nextcloud app skeleton
	cd $(NEXTCLOUD_DIR) && composer install --no-interaction && npm run build

nextcloud-test: nextcloud ## lint + test the Nextcloud app
	cd $(NEXTCLOUD_DIR) && composer lint >/dev/null && composer test && npm test

jellyfin: ## build the Jellyfin plugin (separate GPL repo, checked out next to this one)
	cd $(JELLYFIN_DIR) && dotnet build -c Release

clean: ## remove build output
	rm -rf $(DIST) $(NEXTCLOUD_DIR)/vendor $(NEXTCLOUD_DIR)/node_modules $(NEXTCLOUD_DIR)/js
