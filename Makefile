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
	addlicense -v -check -f ./header.txt ./*.go

build:
	$(GO_BUILD_FLAGS) $(GOBUILD) -o $(BINARY_NAME) -v

test: build
	$(GOTEST) -v ./... 2>&1

integration-test: build
	# temporary workaround, remove next line when x/net false positive is fixed
	echo 'CVE-2018-17142 until=2020-10-22 #x/net false positive\nCVE-2018-17846 until=2020-10-22\nCVE-2018-17143 until=2020-10-22\nCVE-2018-17847 until=2020-10-22\nCVE-2018-17848 until=2020-10-22' > packages/testdata/.nancy-ignore
	cd packages/testdata && GOPATH=. ../../$(BINARY_NAME) sleuth -p Gopkg.lock && cd -
	# temporary workaround, remove next line when x/net false positive is fixed
	mv packages/testdata/.nancy-ignore .
	go list -json -m all | ./$(BINARY_NAME) sleuth
	go list -m all | ./$(BINARY_NAME) sleuth
	go list -json -m all > deps.out && ./$(BINARY_NAME) sleuth < deps.out
	go list -m all > deps.out && ./$(BINARY_NAME) sleuth < deps.out
	# temporary workaround, remove next line when x/net false positive is fixed
	rm .nancy-ignore
