.ONESHELL:
.SILENT:
.DEFAULT_GOAL = help

.PHONY: all
all: clean pre-commit test build ## Clean, run `pre-commit`, test, and build

.PHONY: clean
clean: ## Removes previously created build artifacts
	rm -f ./k6

.PHONY: format
format: ## Formats Go files
	golangci-lint fmt

LINT_OPTIONS = --fix # Fix the lint violations

.PHONY: lint
lint: ## Lints Go files
	golangci-lint run $(LINT_OPTIONS)

.PHONY: test
test: ## Executes unit tests
	go test -cover -race ./...

.PHONY: build
build: ## Builds a custom `k6` binary with the local extension
	go install go.k6.io/xk6/cmd/xk6@latest
	xk6 build --with $(shell go list -m)=.

.PHONY: pre-commit
pre-commit: ## Runs `pre-commit` on all the files
	pre-commit install --install-hooks
	pre-commit run --all-files

.PHONY: help
help: ## Prints help for targets with comments
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%s:\033[0m %s\n", $$1, $$2}' | \
		column -c2 -ts:
