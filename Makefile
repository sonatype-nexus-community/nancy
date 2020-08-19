# Go parameters
GO_BUILD_FLAGS=GO111MODULE=on CGO_ENABLED=0
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=nancy
BUILD_VERSION_LOCATION=github.com/sonatype-nexus-community/nancy/buildversion
GOLANGCI_VERSION=v1.24.0
GOLANGCI_LINT_DOCKER=golangci/golangci-lint:$(GOLANGCI_VERSION)
LINT_CMD=golangci-lint cache status --color always && golangci-lint run --timeout 5m --color always -v --max-same-issues 10

all: deps test lint build

.PHONY: lint clean deps build test integration-test package

lint:
	docker run --rm -v $$(pwd):/app -v $$(pwd)/.cache:/root/.cache -w /app $(GOLANGCI_LINT_DOCKER) /bin/sh -c "$(LINT_CMD)"

ci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_VERSION)
	$(LINT_CMD)

clean:
	$(GOCLEAN)
	rm -rf .cache
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*amd64*

deps:
	$(GOCMD) mod download
	$(GOCMD) mod verify
	$(GOCMD) mod tidy

headers:
	$(GOCMD) get -u github.com/google/addlicense
	addlicense -check -f ./header.txt ./*.go

build: 
	$(GO_BUILD_FLAGS) $(GOBUILD) -o $(BINARY_NAME) -v

test: build
	$(GOTEST) -v ./... 2>&1

integration-test: build
	cd packages/testdata && GOPATH=. ../../$(BINARY_NAME) -p Gopkg.lock && cd -
	go list -m all | ./$(BINARY_NAME)
	go list -m all > deps.out && ./$(BINARY_NAME) < deps.out
