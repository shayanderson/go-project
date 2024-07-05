PROJECT := go-project
DOCKER_NAME := myrepo/$(PROJECT)
DIR_MAIN := ./cmd/app

SOURCES := $(shell find . -name "*.go")
.DEFAULT_GOAL := help
SHELL := /bin/bash
.ONESHELL:

ifneq (,$(wildcard .env))
    include .env
    export
endif
define err
	@echo -e "\033[0;31mERROR: $(1)\033[0m"
	@exit 1
endef

.PHONY: audit
.SILENT: audit
audit: ## Run QC checks (tests, go mod verify, go vet, govulncheck, gosec)
	echo "running audit..."
	go mod verify || exit 1
	go vet ./... || exit 1
	go test -race -vet=off ./...
	if [ $$? -ne 0 ]; then
		$(call err,"tests failed")
	fi
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	if [ $$? -ne 0 ]; then
		$(call err,"vulnerabilities found")
	fi
	go run github.com/securego/gosec/v2/cmd/gosec@latest ./...
	if [ $$? -ne 0 ]; then
		$(call err,"security issues found")
	fi

build: $(SOURCES) ## Build the project
	CGO_ENABLED=0 go build -o bin/$(PROJECT) $(DIR_MAIN)

.PHONY: docker-build
docker-build: ## Build the docker image
	docker build --build-arg PROJECT=$(PROJECT) -t $(DOCKER_NAME) .

.PHONY: help
help: ## Display help
	@MKH_COL_W=20
	@grep -h '##' $(MAKEFILE_LIST) | \
	  grep -v grep | grep -v MKH_COL_W | sort | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-'$$MKH_COL_W's\033[0m %s\n", $$1, $$2}' | \
	  sort

run: build ## Run the project
	./bin/$(PROJECT)

.PHONY: test
.SILENT: test
test: ## Run tests (use `test v=1` to see verbose output)
	echo "running tests..."
	VERBSTR=""
	if [ -n "${v}" ]; then
		VERBSTR=" -v"
	fi
	set -o pipefail ; \
		go test$$VERBSTR -race ./... \
		| sed ''/^ok/s//$$(printf "\033[32mok\033[0m")/'' \
		| sed ''/FAIL/s//$$(printf "\033[31mFAIL\033[0m")/'' \
		| sed ''/\(cached\)/s//$$(printf "\033[33m\(cached\)\033[0m")/''

.PHONY: test-bench
test-bench: ## Run tests with benchmarks
	go test -bench=. -benchmem -cpu=1,2,4 -benchtime=2s -run=NONE ./...

.PHONY: test-cover
.SILENT: test-cover
test-cover: ## Run tests and display coverage
	echo "running tests with coverage..."
	set -o pipefail ; \
		go test -race -coverprofile=/tmp/testcoverage.txt ./... \
		| sed ''/^ok/s//$$(printf "\033[32mok\033[0m")/'' \
		| sed ''/FAIL/s//$$(printf "\033[31mFAIL\033[0m")/''
	echo ---
	go tool cover -func=/tmp/testcoverage.txt | grep total | awk '{print "Total coverage: " $$3}'
	go tool cover -html=/tmp/testcoverage.txt

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy -v

.PHONY: update
update: ## Update dependencies
	go get -u ./...
	go mod tidy -v