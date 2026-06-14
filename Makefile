.PHONY: test lint build fmt vet

fmt:
	gofmt -w ./cmd ./internal

vet:
	go vet ./...

test:
	go test ./...

build:
	go build -o bin/flickr ./cmd/flickr

lint: fmt vet test


changelog:
	git cliff -o CHANGELOG.md

changelog-preview:
	git cliff --latest
