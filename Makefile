.PHONY: build run run-worker start stop test lint clean build-cli

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
RUN_DIR := .run

build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

build-cli:
	go build -o bin/anvil ./cmd/anvil

run: build
	./bin/server

run-worker: build
	./bin/worker

# 后台启动 server / worker，PID 写入 .run/ 便于 stop 精确结束进程。
start: build
	@mkdir -p $(RUN_DIR)
	@if [ -f $(RUN_DIR)/server.pid ] && kill -0 $$(cat $(RUN_DIR)/server.pid) 2>/dev/null; then \
		echo "server already running (pid $$(cat $(RUN_DIR)/server.pid))"; \
	else \
		./bin/server & echo $$! > $(RUN_DIR)/server.pid; \
		echo "started server (pid $$(cat $(RUN_DIR)/server.pid))"; \
	fi
	@if [ -f $(RUN_DIR)/worker.pid ] && kill -0 $$(cat $(RUN_DIR)/worker.pid) 2>/dev/null; then \
		echo "worker already running (pid $$(cat $(RUN_DIR)/worker.pid))"; \
	else \
		./bin/worker & echo $$! > $(RUN_DIR)/worker.pid; \
		echo "started worker (pid $$(cat $(RUN_DIR)/worker.pid))"; \
	fi

# 先按 PID 文件停止，再兜底 pgrep；无进程时不报错。
stop:
	@for role in server worker; do \
		pidfile="$(RUN_DIR)/$$role.pid"; \
		if [ -f "$$pidfile" ]; then \
			pid=$$(cat "$$pidfile"); \
			if kill -0 "$$pid" 2>/dev/null; then \
				kill -TERM "$$pid" 2>/dev/null && echo "stopped $$role (pid $$pid)"; \
			else \
				echo "$$role not running (stale pid $$pid)"; \
			fi; \
			rm -f "$$pidfile"; \
		fi; \
	done
	@for role in server worker; do \
		pids=$$( { pgrep -f "$(ROOT_DIR)/bin/$$role" 2>/dev/null; pgrep -f "^\\./bin/$$role\$$" 2>/dev/null; } | sort -u); \
		if [ -n "$$pids" ]; then \
			echo "stopping leftover bin/$$role: $$pids"; \
			kill -TERM $$pids 2>/dev/null || true; \
		fi; \
	done; \
	true

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin var/log/anvil $(RUN_DIR)
