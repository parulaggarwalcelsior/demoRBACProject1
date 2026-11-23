SHELL := /bin/sh

TOOLS_DIR := $(CURDIR)/.tools/bin
GOIMPORTS_VERSION := v0.31.0
GOLANGCI_LINT_VERSION := v1.64.8

GOIMPORTS := $(TOOLS_DIR)/goimports
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint

GO_PACKAGES := ./...
FORMAT_DIRS := internal tests

.PHONY: tools schema-check fmt-check gofmt-check goimports-check vet lint race-test clean-tools

tools: $(GOIMPORTS) $(GOLANGCI_LINT)

$(GOIMPORTS):
	@mkdir -p "$(TOOLS_DIR)"
	GOBIN="$(TOOLS_DIR)" go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

$(GOLANGCI_LINT):
	@mkdir -p "$(TOOLS_DIR)"
	GOBIN="$(TOOLS_DIR)" go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

schema-check: fmt-check vet lint race-test

fmt-check: gofmt-check goimports-check

gofmt-check:
	@files="$$(gofmt -l $(FORMAT_DIRS))"; \
	if [ -n "$$files" ]; then \
		printf '%s\n' "$$files"; \
		exit 1; \
	fi

goimports-check: $(GOIMPORTS)
	@files="$$("$(GOIMPORTS)" -l $(FORMAT_DIRS))"; \
	if [ -n "$$files" ]; then \
		printf '%s\n' "$$files"; \
		exit 1; \
	fi

vet:
	go vet $(GO_PACKAGES)

lint: $(GOLANGCI_LINT)
	"$(GOLANGCI_LINT)" run

race-test:
	go test -race $(GO_PACKAGES) -count=1

clean-tools:
	rm -rf "$(CURDIR)/.tools"
