SHELL := /bin/bash

EXCLUDED_PACKAGES := \
	github.com/foxboron/terraform-provider-openwrt \
	github.com/foxboron/terraform-provider-openwrt/internal/api \
	github.com/foxboron/terraform-provider-openwrt/internal/types \
	github.com/foxboron/terraform-provider-openwrt/mocks

PACKAGES := $(shell go list ./... | grep -Fvx -f <(printf '%s\n' $(EXCLUDED_PACKAGES)))

default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

bin-deps:
	go install go.uber.org/mock/mockgen@latest

test:
ifeq ($(TEST_PACKAGE),)
	go test -parallel=10 -timeout 120s -cover -coverprofile=_out/.coverage -v ./...;
else
	go test -parallel=10 -timeout 120s -cover -coverprofile=_out/.coverage -v ./$(TEST_PACKAGE);
endif
	go tool cover -html=_out/.coverage;

testacc:
	TF_ACC=1 go test -timeout 20s -cover -coverprofile=_out/.coverage -v $(PACKAGES)

coveragefile:
	go tool cover -html=_out/.coverage -o=./_out/coverage.html

release:
	export GITHUB_TOKEN=$(shell gh config get oauth_token -h github.com) && \
	export GPG_FINGERPRINT="9C02FF419FECBE16" && \
	goreleaser release --clean

.PHONY: bin-deps fmt lint test testacc build install generate release
