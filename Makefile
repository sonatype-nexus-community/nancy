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
NANCY_IGNORE=$(shell cat .nancy-ignore | cut -d\# -f 1)
IT_EXCLUDED_VULNS=CVE-2021-3121,CVE-2022-21698,CVE-2022-29153,sonatype-2021-1401,CVE-2023-32731,CVE-2023-45142,CVE-2024-10086

ifeq ($(findstring localbuild,$(CIRCLE_SHELL_ENV)),localbuild)
    DOCKER_CMD=sudo docker
else
    DOCKER_CMD=docker
endif

all: deps test lint build

.PHONY: lint clean deps build test integration-test package

lint:
	$(DOCKER_CMD) run --rm -v $$(pwd):/app -v $$(pwd)/.cache:/root/.cache -w /app $(GOLANGCI_LINT_DOCKER) /bin/sh -c "$(LINT_CMD)"

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
	$(GOCMD) mod tidy -compat=1.17

headers:
	$(GOCMD) get -u github.com/google/addlicense
	addlicense -check -f ./header.txt ./*.go

build:
	$(GO_BUILD_FLAGS) $(GOBUILD) -o $(BINARY_NAME) -v

test: build
	$(GOTEST) -v 2>&1 ./...

integration-test: build
	mkdir -p dist
	cd packages/testdata && GOPATH=. ../../$(BINARY_NAME) sleuth -p Gopkg.lock && cd -
	go list -json -deps ./... | ./$(BINARY_NAME) sleuth
	go list -json -deps | ./$(BINARY_NAME) sleuth
	go list -json -m all | ./$(BINARY_NAME) sleuth --exclude-vulnerability $(IT_EXCLUDED_VULNS)
	go list -m all | ./$(BINARY_NAME) sleuth --exclude-vulnerability $(IT_EXCLUDED_VULNS)
	go list -json -deps ./... > dist/deps.out && ./$(BINARY_NAME) sleuth < dist/deps.out
	go list -json -deps > dist/deps.out && ./$(BINARY_NAME) sleuth < dist/deps.out
	go list -json -m all > dist/deps.out && ./$(BINARY_NAME) sleuth --exclude-vulnerability $(IT_EXCLUDED_VULNS) < dist/deps.out
	go list -m all > dist/deps.out && ./$(BINARY_NAME) sleuth --exclude-vulnerability $(IT_EXCLUDED_VULNS) < dist/deps.out

build-linux:
	GOOS=linux GOARCH=amd64 $(GO_BUILD_FLAGS) $(GOBUILD) -o $(BINARY_NAME) -v

docker-alpine-integration-test: build-linux
	mkdir -p dist
	$(DOCKER_CMD) build . -f Dockerfile.alpine -t sonatypecommunity/nancy:alpine-integration-test
	# create deps output since this container does not have golang,
    # It's simulating the following flow
    # 1. ci runs using a container that has golang, it then exports the go list -json m all contents
    # 2. passes it to the next step that is using this container that only has nancy in it
    # 3. runs nancy using the contents of the exported file with the deps in it. Also assumes that
    #    in ci its likely you have the codebase (thus .nancy-ignore) in the same location you run nancy sleuth
	go list -json -deps ./... > dist/deps.out
	echo "cd /tmp && cat /tmp/dist/deps.out | nancy sleuth" > dist/ci.sh
	echo "cd /tmp && cat /tmp/dist/deps.out | nancy sleuth --output=json && > nancy-result.json && cat nancy-result.json | jq '.'" > dist/ci-json.sh
	chmod +x dist/ci.sh
	chmod +x dist/ci-json.sh
	# run the container....using cat with no params keeps it running
	$(DOCKER_CMD) run --name alpine-integration-test -td sonatypecommunity/nancy:alpine-integration-test cat
	# copy the code as if it was actually in the "ci" container.. doing this cause circleci cant actually mount volumes
	$(DOCKER_CMD) cp . alpine-integration-test:/tmp
	# run nancy against nancy output
	$(DOCKER_CMD) exec -it alpine-integration-test /bin/sh /tmp/dist/ci.sh
	$(DOCKER_CMD) exec -it alpine-integration-test /bin/sh /tmp/dist/ci-json.sh
	$(DOCKER_CMD) stop alpine-integration-test && $(DOCKER_CMD) rm alpine-integration-test


docker-goreleaser-integration-test: build-linux
	$(DOCKER_CMD) build . -f Dockerfile.goreleaser -t sonatypecommunity/nancy:goreleaser-integration-test
	# NANCY_IGNORE is more tomfoolery b/c circleci cant do volume mounts. Use the non-file ignore version but with the contents of
    # the .nancy-ignore. If you were to do this for real you would likely volume mount to your local and it
    # would just use whatever file you actually had.
	go list -json -deps ./... | $(DOCKER_CMD) run --rm -i sonatypecommunity/nancy:goreleaser-integration-test sleuth -e $(NANCY_IGNORE)

docker-integration-tests: docker-alpine-integration-test docker-goreleaser-integration-test

goreleaser-sanity-check:
	# verifies goreleaser can build all the stuff. couldn't get it to play nice in the regular CI 'build' job, but at
	# the incantation is here for future use.
	curl -sL https://git.io/goreleaser | bash -s -- --snapshot --skip-publish --rm-dist
