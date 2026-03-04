.PHONY: build clean test

BINARY := wellspring
BUILD_DIR := build
LDFLAGS := -s -w -extldflags '-static'

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/wellspring/

clean:
	rm -rf $(BUILD_DIR)

test:
	go test -race -count=1 ./...
