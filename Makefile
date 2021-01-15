PHONY: generate
generate:
	go mod tidy
	go generate ./...
