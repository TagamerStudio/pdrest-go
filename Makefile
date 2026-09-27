.PHONY: test lint vet fmt fuzz tidy check cover help

GOLANGCI_LINT_VERSION ?= v2.12.2
GOLANGCI_LINT_CMD = go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
GOFUMPT_VERSION ?= v0.7.0
GOIMPORTS_VERSION ?= v0.30.0
FUZZ_TARGET ?= FuzzNormalizeBaseURL

test:
	go test -count=1 -race -v ./...

cover:
	go test -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep -E "total:.*100\.0%"

lint:
	$(GOLANGCI_LINT_CMD) run --config .golangci.yml --timeout 5m ./...

vet:
	go vet ./...

fuzz:
	go test -run=^$$ -fuzz=$(FUZZ_TARGET) -fuzztime=30s .

fmt:
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	gofmt -s -w .
	$(shell go env GOPATH)/bin/goimports -w .
	$(shell go env GOPATH)/bin/gofumpt -w .

tidy:
	go mod tidy

check: vet lint test cover

help:
	@echo 'Targets: test lint vet fmt fuzz tidy check cover help'
