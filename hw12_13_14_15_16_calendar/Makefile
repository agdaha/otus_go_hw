BIN := "./bin/calendar"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_SENDER := "./bin/calendar_sender"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

#Postrges
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB ?= backend
POSTGRES_PORT ?= 5435
POSTGRES_CONTAINER := postgres-calendar

#Rabbit
RABBIT_USER ?= rabbit
RABBIT_PASSWORD ?= password
RABBIT_PORT ?= 5672
RABBIT_PORT_UI ?= 15672
RABBIT_CONTAINER := rabbitmq


run-postgres:
	docker run -d --name $(POSTGRES_CONTAINER) \
	-e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	-e POSTGRES_DB=$(POSTGRES_DB) \
	-p $(POSTGRES_PORT):5432 \
	-v postgres-data:/var/lib/postgresql/data \
	postgres:latest

run-rabbit:
	docker run -d --name $(RABBIT_CONTAINER) \
	-p $(RABBIT_PORT):5672 \
	-p $(RABBIT_PORT_UI):15672 \
	-e RABBITMQ_DEFAULT_USER=$(RABBIT_USER) \
	-e RABBITMQ_DEFAULT_PASS=$(RABBIT_PASSWORD) \
	rabbitmq:management

#Protobuf
PROTOC_VERSION := 25.5
PROTOC_GEN_GO_VERSION := v1.36.1
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1
GOBIN := $(shell go env GOPATH)/bin
PROTOC_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/osx/')
PROTOC_ARCH := $(shell uname -m | sed 's/arm64/aarch_64/;s/aarch64/aarch_64/')

install-proto-deps:
	@which protoc > /dev/null || ( \
		echo "installing protoc $(PROTOC_VERSION)..." && \
		curl -sSfL -o /tmp/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip && \
		unzip -oq /tmp/protoc.zip -d /tmp/protoc-install && \
		install -m 755 /tmp/protoc-install/bin/protoc $(GOBIN)/protoc && \
		rm -rf /tmp/protoc.zip /tmp/protoc-install \
	)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

generate: install-proto-deps
	PATH="$(GOBIN):$$PATH" go generate ./...

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

run: build
	$(BIN) -config ./configs/config.yaml

run-scheduler: build
	$(BIN_SCHEDULER) -config ./configs/scheduler_config.yaml

run-sender: build
	$(BIN_SENDER) -config ./configs/sender_config.yaml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.57.2

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: install-proto-deps generate build run run-scheduler run-sender build-img run-img version test lint
