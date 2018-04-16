.DEFAULT_GOAL := test

setup:
	dep ensure

fmt:
	go fmt

get:
	go vet

test:
	go test -race ./...

coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
