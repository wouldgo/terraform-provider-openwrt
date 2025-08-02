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

test:
	rm -Rf _out/.coverage;
ifeq ($(TEST_PACKAGE),)
	go test -parallel=10 -timeout 120s -cover -coverprofile=_out/.coverage -v ./...;
else
	go test -parallel=10 -timeout 120s -cover -coverprofile=_out/.coverage -v ./$(TEST_PACKAGE);
endif
	go tool cover -html=_out/.coverage;

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

release:
	export GITHUB_TOKEN=$(shell gh config get oauth_token -h github.com) && \
	export GPG_FINGERPRINT="9C02FF419FECBE16" && \
	goreleaser release --clean

.PHONY: fmt lint test testacc build install generate release
