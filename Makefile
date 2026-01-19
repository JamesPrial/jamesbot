.PHONY: build run test clean fmt

build:
	go build -o bin/jamesbot ./cmd/bot

run: build
	./bin/jamesbot

test:
	go test -v -race ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin/
