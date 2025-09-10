# The short Git commit hash
SHORT_COMMIT := $(shell git rev-parse --short HEAD)
# The Git commit hash
COMMIT := $(shell git rev-parse HEAD)
# The tag of the current commit, otherwise empty
VERSION := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
# Name of the cover profile
COVER_PROFILE := coverage.txt
# Disable go sum database lookup for private repos
GOPRIVATE := github.com/dapperlabs/*
# Ensure go bin path is in path (Especially for CI)
GOPATH ?= $(HOME)/go
PATH := $(PATH):$(GOPATH)/bin
# OS
UNAME := $(shell uname)

MIXPANEL_PROJECT_TOKEN := 3fae49de272be1ceb8cf34119f747073

.PHONY: test
test:
	GO111MODULE=on go test -v -coverprofile=$(COVER_PROFILE) $(if $(JSON_OUTPUT),-json,) ./...

.PHONY: install-tools
install-tools:
	cd ${GOPATH}; \
	mkdir -p ${GOPATH}; \
	GO111MODULE=on go install github.com/axw/gocov/gocov@latest; \
	GO111MODULE=on go install github.com/matm/gocov-html/cmd/gocov-html@latest; \
	GO111MODULE=on go install github.com/sanderhahn/gozip/cmd/gozip@latest; \
	GO111MODULE=on go install github.com/vektra/mockery/v2@v2.53.5;

.PHONY: generate-schema
generate-schema:
	go run ./cmd/flow-schema/flow-schema.go ./schema.json

.PHONY: check-schema
check-schema:
	go run ./cmd/flow-schema/flow-schema.go --verify=true ./schema.json

.PHONY: check-tidy
check-tidy:
	go mod tidy

.PHONY: check-headers
check-headers:
	@./check-headers.sh

.PHONY: generate
generate: install-tools
	go generate ./...

.PHONY: coverage
coverage:
ifeq ($(COVER), true)
	# file has to be called index.html
	gocov convert $(COVER_PROFILE) > cover.json
	./cover-summary.sh
	gocov-html cover.json > index.html
	# coverage.zip will automatically be picked up by teamcity
	gozip -c coverage.zip index.html
endif
