.PHONY: build run test clean fmt lint

build:
	go build -o bin/jamesbot ./cmd/bot

run: build
	./bin/jamesbot

test:
	go test -v -race ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
