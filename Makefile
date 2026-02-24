PLUGIN_ID = se.bylund.mattermost-plugin-rc-migrate
PLUGIN_VERSION = 1.1.0
BUNDLE_NAME = $(PLUGIN_ID)-$(PLUGIN_VERSION).tar.gz

GO ?= go
GOFLAGS ?= -mod=vendor

# Build targets matching plugin.json executables
TARGETS = linux-amd64 linux-arm64 darwin-amd64 darwin-arm64

.PHONY: build clean vendor dist

build: dist/$(BUNDLE_NAME)

dist/$(BUNDLE_NAME): server-build
	@echo "Bundling plugin..."
	@mkdir -p dist
	@cd dist/intermediate && tar czf ../$(BUNDLE_NAME) $(PLUGIN_ID)
	@echo "Built: dist/$(BUNDLE_NAME)"

server-build:
	@echo "Building server components..."
	@mkdir -p dist/intermediate/$(PLUGIN_ID)
	@cp plugin.json dist/intermediate/$(PLUGIN_ID)/
	@for target in $(TARGETS); do \
		os=$$(echo $$target | cut -d- -f1); \
		arch=$$(echo $$target | cut -d- -f2); \
		echo "  Building $$os/$$arch..."; \
		mkdir -p dist/intermediate/$(PLUGIN_ID)/server/dist; \
		GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) -o dist/intermediate/$(PLUGIN_ID)/server/dist/plugin-$$target ./server; \
	done

vendor:
	$(GO) mod vendor

clean:
	rm -rf dist
