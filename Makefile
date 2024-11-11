
NAME=woaa
BUILD_DIR=build
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags '-w -s -buildid='

normal: clean default

all: clean all-platform

clean:
	rm -rf $(BUILD_DIR)

default:
	$(GOBUILD) -o $(BUILD_DIR)/$(NAME)

all-platform: linux-amd64 linux-arm64 windows-amd64 darwin-amd64 darwin-arm64

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-$@

darwin-arm64:
	GOARCH=arm64 GOOS=darwin $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-$@

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BUILD_DIR)/$(NAME)-$@.exe

test:
	go test -v ./...

run:
	go build -o $(BUILD_DIR)/woaa main.go
	cp .env $(BUILD_DIR)/.env
	cd $(BUILD_DIR) && ./woaa
