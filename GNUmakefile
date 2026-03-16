default: build

build:
	go build -o terraform-provider-fly

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/stategraph/fly/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-fly ~/.terraform.d/plugins/registry.terraform.io/stategraph/fly/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./internal/... -v -parallel 2 -timeout 30m

sweep:
	go test ./internal/provider/... -v -sweep=all -timeout 15m

lint:
	golangci-lint run ./...

generate:
	go generate ./...

.PHONY: build install test testacc sweep lint generate
