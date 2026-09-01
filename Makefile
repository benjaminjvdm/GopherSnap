.PHONY: all build test clean release

BINARY_NAME=gophersnap
WINDOWS_BINARY=gophersnap.exe

all: build

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) main.go

test:
	go test -v -cover ./...

clean:
	rm -f $(BINARY_NAME) $(WINDOWS_BINARY) gophersnap_*.zip gophersnap_*.tar.gz coverage.out

release:
	go run github.com/tc-hib/go-winres@latest make --arch amd64,arm64,386
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(WINDOWS_BINARY) main.go
	zip -q gophersnap_windows_x64.zip $(WINDOWS_BINARY) README.md LICENSE
