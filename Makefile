# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=nancy
BUILD_VERSION_LOCATION=github.com/sonatype-nexus-community/nancy/buildversion
GOLANGCI_LINT_DOCKER=golangci/golangci-lint:v1.23.1

all: deps test lint build

.PHONY: lint clean deps env-setup build test integration-test package

lint:
	docker run --rm -v $$(pwd):/app -v $$(pwd)/.cache:/root/.cache -w /app $(GOLANGCI_LINT_DOCKER) /bin/sh -c "golangci-lint cache status --color always && golangci-lint run --timeout 5m --color always -v"

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*amd64*

deps:
	$(GOCMD) mod download
	$(GOCMD) mod verify
	$(GOCMD) mod tidy

env-setup:
	export GO111MODULE=on CGO_ENABLED=0

build: env-setup
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

integration-test: env-setup build
	cd testdata/dep && ../../$(BINARY_NAME) Gopkg.lock && cd -
	./$(BINARY_NAME) go.sum
	$(GOCMD) list -m all | ./$(BINARY_NAME)

package: env-setup
	export VERSION=$(git describe --abbrev=0 --tags) && export LAST_PREFIX=$(cut -d'.' -f1,2 <<< $VERSION) && export LAST_SUFFIX=$(cut -d'.' -f3 <<< $VERSION) && export NEW_SUFFIX=$(expr "$LAST_SUFFIX" + 1) && export VERSION="$LAST_PREFIX.$NEW_SUFFIX"
	echo $(VERSION)
	GOARCH=amd64
	GOOS=linux $(GOBUILD) -ldflags="-X '$(BUILD_VERSION_LOCATION).BuildVersion=$(VERSION)' -X '$(BUILD_VERSION_LOCATION).BuildTime=$(time)' -X '$(BUILD_VERSION_LOCATION).BuildCommit=$(TRAVIS_COMMIT)'" -o $(BINARY_NAME)-linux.amd64-$(VERSION)
	GOOS=darwin $(GOBUILD) -ldflags="-X '$(BUILD_VERSION_LOCATION).BuildVersion=$(VERSION)' -X '$(BUILD_VERSION_LOCATION).BuildTime=$(time)' -X '$(BUILD_VERSION_LOCATION).BuildCommit=$(TRAVIS_COMMIT)'" -o $(BINARY_NAME)-darwin.amd64-$(VERSION)
	GOOS=windows $(GOBUILD) -ldflags="-X '$(BUILD_VERSION_LOCATION).BuildVersion=$(VERSION)' -X '$(BUILD_VERSION_LOCATION).BuildTime=$(time)' -X '$(BUILD_VERSION_LOCATION).BuildCommit=$(TRAVIS_COMMIT)'" -o $(BINARY_NAME)-windows.amd64-$(VERSION).exe
