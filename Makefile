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
	mkdir -p dist
	cd packages/testdata && GOPATH=. ../../$(BINARY_NAME) sleuth -p Gopkg.lock && cd -
	go list -json -m all | ./$(BINARY_NAME) sleuth
	go list -m all | ./$(BINARY_NAME) sleuth
	go list -json -m all > dist/deps.out && ./$(BINARY_NAME) sleuth < dist/deps.out
	go list -m all > dist/deps.out && ./$(BINARY_NAME) sleuth < dist/deps.out

build-linux:
	GOOS=linux GOARCH=amd64 $(GO_BUILD_FLAGS) $(GOBUILD) -o $(BINARY_NAME) -v

docker-alpine-integration-test: build-linux
	mkdir -p dist
	docker build . -f Dockerfile.alpine -t sonatypecommunity/nancy:alpine-integration-test
	# create file, volume mount to simulate, ci run of the container and things just happening inside the container instead of passing output to the container directly
	go list -json -m all > dist/deps.out && docker run -v $$(pwd):/go/src/github.com/user/repo -it --rm sonatypecommunity/nancy:alpine-integration-test cat /go/src/github.com/user/repo/dist/deps.out | nancy sleuth

docker-goreleaser-integration-test: build-linux
	docker build . -f Dockerfile.goreleaser -t sonatypecommunity/nancy:goreleaser-integration-test
	go list -json -m all | docker run --rm -i sonatypecommunity/nancy:goreleaser-integration-test sleuth

docker-integration-tests: docker-alpine-integration-test docker-goreleaser-integration-test
