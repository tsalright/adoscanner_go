# The name of the executable (default is current directory name)
TARGET := "adoscanner"
APP_VERSION := 1.0.0
.DEFAULT_GOAL := help

COVER_FILE := coverage.txt

# Go parameters
GOCMD=go
GOPATH:=$(shell $(GOCMD) env GOPATH)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOTEST=$(GOCMD) test -race
GOLINT=$(GOPATH)/bin/golangci-lint
GO2XUNIT=$(GOPATH)/bin/go2xunit
GOCOV=$(GOPATH)/bin/gocov
GOCOVXML=$(GOPATH)/bin/gocov-xml

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

.PHONY: all lint clean

all: test lint build ## runs the test, lint and build targets

check-go:
ifndef GOPATH
	$(error "go is not available please install golang version 1.13+, https://golang.org/dl/")
endif

build: check-go ## builds
	env CGO_ENABLED=0 GOOS=linux $(GOBUILD) .

ci-build: clean lint ci-test ci-cover build

ci-test: check-go
	$(GOGET) -u github.com/tebeka/go2xunit
	$(GOTEST) ./... | $(GO2XUNIT) > test_output.xml

test: check-go ## recursively tests all .go files
	$(GOTEST) ./...

clean: check-go ## runs `go clean`
	$(GOCLEAN) --modcache
	rm -f $(TARGET)
	rm -f $(COVER_FILE)
	rm -f coverage.xml
	rm -f test_output.xml

ci-cover: check-go ## runs test suite with coverage profiling
	$(GOGET) github.com/axw/gocov/gocov
	$(GOGET) github.com/AlekSi/gocov-xml
	$(GOTEST) ./... -coverprofile=$(COVER_FILE)
	$(GOCOV) convert $(COVER_FILE) | $(GOCOVXML) > coverage.xml

cover: check-go ## runs test suite with coverage profiling
	$(GOTEST) ./... -coverprofile=$(COVER_FILE)

lint: check-go ## runs `golangci-lint` linters defined in `.golangci.yml` file
ifeq (,$(wildcard $(GOLINT)))
	echo $(GOPATH)/bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.24.0
endif
	$(GOLINT) run --out-format=tab --tests=false ./...

run: build ## builds and executes the TARGET binary
	./$(TARGET)

help: ## help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
