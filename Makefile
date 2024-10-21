MAKEFLAGS += --silent

all: clean pre-commit build

help: ## Prints help for targets with comments
	@cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

clean: ## Removes previously created build artifacts.
	rm -f ./k6

build: ## Builds a custom `k6` binary with the local extension.
	go install go.k6.io/xk6/cmd/xk6@latest
	xk6 build --with $(shell go list -m)=.

pre-commit: ## Runs `pre-commit` on all the files.
	pre-commit run --all-files

format: ## Formats Go files.
	go fmt ./...

lint: ## Lints Go files.
	golangci-lint run --fix

test: ## Executes unit tests.
	go test -cover -race ./...

.PHONY: build clean format help pre-commit test
