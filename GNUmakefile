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
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

release:
	GITHUB_TOKEN=$(gh config get oauth_token -h github.com) GPG_FINGERPRINT="9C02FF419FECBE16" goreleaser release --clean

.PHONY: fmt lint test testacc build install generate release
