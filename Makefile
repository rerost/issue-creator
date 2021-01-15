PHONY: setup
setup:
	GO111MODULE=off go get github.com/izumin5210/gex/cmd/gex
	gex --build

PHONY: generate
generate: setup
	go mod tidy
	go generate ./...
