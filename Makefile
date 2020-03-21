# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=nancy
BUILD_VERSION_LOCATION=github.com/sonatype-nexus-community/nancy/buildversion
GOLANGCI_VERSION=v1.23.6
GOLANGCI_LINT_DOCKER=golangci/golangci-lint:$(GOLANGCI_VERSION)
LINT_CMD=golangci-lint cache status --color always && golangci-lint run --timeout 5m --color always -v --max-same-issues 10

all: deps test lint build

.PHONY: lint clean deps env-setup build test integration-test package

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

env-setup:
	export GO111MODULE=on CGO_ENABLED=0

build: env-setup
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -race -v -count=1 -p=1 ./... 2>&1

integration-test: env-setup build
	cd packages/testdata && ../../$(BINARY_NAME) Gopkg.lock && cd -
	./$(BINARY_NAME) go.sum
	go list -m all | ./$(BINARY_NAME)
	go list -m all > deps.out && ./$(BINARY_NAME) < deps.out
