.PHONY: build run run-worker test lint clean

build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

# 最小 HTTP 启动只构建 server，避免 worker 构建问题阻塞本地起服务。
run:
	go build -o bin/server ./cmd/server
	./bin/server

run-worker:
	go build -o bin/worker ./cmd/worker
	./bin/worker

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin var/log/anvil-scaffold-project
