APP_NAME ?= app
CMD_PATH ?= ./cmd/main.go
DIST_DIR ?= ./dist

PLATFORMS ?= linux/amd64 linux/arm64 linux/386 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 freebsd/amd64 freebsd/arm64

.PHONY: build build-all run clean

build:
	@mkdir -p $(DIST_DIR)
	@set -e; \
	os=$$(go env GOOS); \
	arch=$$(go env GOARCH); \
	output="$(DIST_DIR)/$(APP_NAME)-$$os-$$arch"; \
	if [ "$$os" = "windows" ]; then output="$$output.exe"; fi; \
	echo "Building $$output"; \
	CGO_ENABLED=0 go build -o "$$output" $(CMD_PATH)

build-all: clean
	@mkdir -p $(DIST_DIR)
	@set -e; \
	for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output="$(DIST_DIR)/$(APP_NAME)-$${os}-$${arch}"; \
		if [ "$$os" = "windows" ]; then output="$$output.exe"; fi; \
		echo "Building $$output"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -o "$$output" $(CMD_PATH); \
	done

run:
	go run $(CMD_PATH)

clean:
	rm -rf $(DIST_DIR)
