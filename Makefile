.PHONY: all build test clean release

BIN_DIR=binaries
BINARY_NAME=$(BIN_DIR)/gophersnap
WINDOWS_BINARY=$(BIN_DIR)/gophersnap.exe

all: build

build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) main.go

test:
	go test -v -cover ./...

clean:
	rm -rf $(BIN_DIR) coverage.out

release:
	mkdir -p $(BIN_DIR)
	go run github.com/tc-hib/go-winres@latest make --arch amd64,arm64,386
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(WINDOWS_BINARY) main.go
	zip -q $(BIN_DIR)/gophersnap_windows_x64.zip $(WINDOWS_BINARY) README.md LICENSE
