default: build
.PHONY: build test testacc docs local acc-up acc-down acc-test acc

PROJ := technitium
ORG := kevynb
TECHNITIUM_API_URL ?= http://localhost:5380
TECHNITIUM_ADMIN_USER ?= admin
TECHNITIUM_ADMIN_PASSWORD ?= changeme
TECHNITIUM_SKIP_TLS_VERIFY ?=

BINARY := terraform-provider-$(PROJ)
VERSION := $(shell git describe --tags --always)

ARCH := $(shell go env GOARCH)
OS := $(shell go env GOOS)

LOCAL_PATH := ~/.terraform.d/plugins/registry.terraform.io/$(ORG)/$(PROJ)/$(VERSION)/$(OS)_$(ARCH)/

export

## cmds

build:
	go build -o bin/$(BINARY) -ldflags='-s -w -X main.version=$(VERSION)' .

test:
	go test -v -timeout=30s -parallel=4 ./...

testacc:
	TF_ACC=1 go test -v -timeout 2m ./...

acc-up:
	docker compose -f tools/acceptance/docker-compose.yml up -d
	TECHNITIUM_API_URL="$(TECHNITIUM_API_URL)" \
	TECHNITIUM_ADMIN_USER="$(TECHNITIUM_ADMIN_USER)" \
	TECHNITIUM_ADMIN_PASSWORD="$(TECHNITIUM_ADMIN_PASSWORD)" \
	TECHNITIUM_SKIP_TLS_VERIFY="$(TECHNITIUM_SKIP_TLS_VERIFY)" \
	tools/acceptance/get-token.sh > /dev/null

acc-down:
	docker compose -f tools/acceptance/docker-compose.yml down -v

acc-test:
	@test -f tools/acceptance/token.env || \
	TECHNITIUM_API_URL="$(TECHNITIUM_API_URL)" \
	TECHNITIUM_ADMIN_USER="$(TECHNITIUM_ADMIN_USER)" \
	TECHNITIUM_ADMIN_PASSWORD="$(TECHNITIUM_ADMIN_PASSWORD)" \
	TECHNITIUM_SKIP_TLS_VERIFY="$(TECHNITIUM_SKIP_TLS_VERIFY)" \
	tools/acceptance/get-token.sh > /dev/null
	@. tools/acceptance/token.env; \
	TECHNITIUM_API_URL="$(TECHNITIUM_API_URL)" \
	TECHNITIUM_SKIP_TLS_VERIFY="$(TECHNITIUM_SKIP_TLS_VERIFY)" \
	TF_ACC=1 \
	go test -v -timeout 10m ./internal/provider -run TestAcc

acc: acc-up acc-test

local: build
	go build -o $(BINARY) -ldflags='-s -w -X main.version=$(VERSION)' .
	rm -rf       $(LOCAL_PATH)
	mkdir -p     $(LOCAL_PATH)
	mv $(BINARY) $(LOCAL_PATH)
	chmod +x     $(LOCAL_PATH)/$(BINARY)
