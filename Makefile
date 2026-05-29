.PHONY: build run run-worker test lint clean build-cli

build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

build-cli:
	go build -o bin/anvil ./cmd/anvil

run: build
	./bin/server

run-worker: build
	./bin/worker

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin var/log/anvil
