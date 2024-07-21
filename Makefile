BIN_DIR := ${PWD}/bin
export PATH := ${BIN_DIR}:${PATH}

PHONY: setup
setup:
	go install github.com/izumin5210/gex/cmd/gex@v0.6.1
	gex --build

PHONY: generate
generate: setup
	go mod tidy
	go generate ./...
