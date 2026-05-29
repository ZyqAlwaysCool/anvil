# syntax=docker/dockerfile:1

ARG REGISTRY=swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library

FROM ${REGISTRY}/golang:1.24.9-alpine AS builder
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

RUN rm -rf /app/bin/
RUN go build -ldflags="-s -w" -o ./bin/server ./cmd/server
RUN go build -ldflags="-s -w" -o ./bin/worker ./cmd/worker
RUN cp -r configs prompts /app/bin/

FROM ${REGISTRY}/alpine:3.18
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk add --no-cache tzdata \
	&& cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
	&& echo "Asia/Shanghai" > /etc/timezone \
	&& apk del tzdata

ARG APP_ENV=prod
ENV APP_ENV=${APP_ENV}

WORKDIR /app
COPY --from=builder /app/bin /app
COPY --from=builder /app/configs/.env.example /app/.env

EXPOSE 28888
ENTRYPOINT ["./server"]
