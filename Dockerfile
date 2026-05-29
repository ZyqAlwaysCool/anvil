# syntax=docker/dockerfile:1

ARG REGISTRY=swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library

FROM ${REGISTRY}/golang:1.24.9-alpine AS builder
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

WORKDIR /app

# 须在 go mod download 之前设置，否则默认走 proxy.golang.org 易超时
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN rm -rf /app/bin/
RUN go build -ldflags="-s -w" -o ./bin/server ./cmd/server
RUN go build -ldflags="-s -w" -o ./bin/worker ./cmd/worker
RUN cp -r configs prompts /app/bin/

FROM ${REGISTRY}/alpine:3.18
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

# 须保留 tzdata：Go 的 time.LoadLocation(LOG_TIMEZONE) 依赖 /usr/share/zoneinfo，仅写 /etc/localtime 不够
RUN apk add --no-cache tzdata \
	&& cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
	&& echo "Asia/Shanghai" > /etc/timezone
ENV TZ=Asia/Shanghai

ARG APP_ENV=prod
ENV APP_ENV=${APP_ENV}

WORKDIR /app
COPY --from=builder /app/bin /app
COPY --from=builder /app/configs/.env.example /app/.env

EXPOSE 28888
CMD ["./server"]
