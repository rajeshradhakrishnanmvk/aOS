.PHONY: build install test lint clean

BINARY_NAME=kubectl-brain
GOPATH?=$(shell go env GOPATH)

build:
	go build -o $(BINARY_NAME) ./cmd/brain/

install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

test:
	go test ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

clean:
	rm -f $(BINARY_NAME)
