.DEFAULT_GOAL := test

guardcmd-%:
	@hash $(*) > /dev/null 2>&1 || \
		(echo "ERROR: '$(*)' must be installed and available on your PATH."; exit 1)

guardenv-%:
	@if [ -z '${${*}}' ]; then echo 'ERROR: environment variable $* not set' && exit 1; fi

fmt:
	@go fmt ./...

lint: guardcmd-gofmt
	@test -z $$(gofmt -d -l . | tee /dev/stderr) && echo "gofmt ok"

vet:
	@go vet

test:
ifdef CI
	LOG_LEVEL=fatal go test -race ./...
else
	@LOG_LEVEL=fatal ginkgo -r
endif

check: vet lint test

coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
