ifndef VERSION
	VERSION = ${shell git rev-parse --abbrev-ref HEAD}-${shell git rev-parse --short HEAD}

endif

ifeq (,$(shell go env GOBIN))
	GOBIN=$(shell go env GOPATH)/bin
else
	GOBIN=$(shell go env GOBIN)
endif
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
GOLANG_CI_LINT = $(LOCALBIN)/golangci-lint
GOLANG_CI_LINT_VERSION ?= v1.60.3

lint:
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)
	$(LOCALBIN)/golangci-lint run --verbose -c .golangci.yaml

FLAGS = -ldflags '-s -w -X github.com/kyma-project/modulectl/cmd/modulectl/version.Version=$(VERSION)'

validate-docs:
	./hack/verify-generated-docs.sh

build: build-linux build-darwin build-linux-arm build-darwin-arm

build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/modulectl-darwin $(FLAGS) ./cmd

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/modulectl-linux $(FLAGS) ./cmd

build-darwin-arm:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/modulectl-darwin-arm $(FLAGS) ./cmd

build-linux-arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/modulectl-linux-arm $(FLAGS) ./cmd

.PHONY: docs
docs:
	rm -f ./docs/gen-docs/*
	go run ./scripts/gendocs/gendocs.go

test:
	go test `go list ./... | grep -v /tests/e2e` -race -coverprofile=cover.out
	@echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')"
	@rm cover.out
