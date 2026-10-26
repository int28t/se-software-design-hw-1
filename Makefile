.DEFAULT_GOAL := check

# Сохраняем переданные флаги и отключаем обращение Go к системе контроля версий.
export GOFLAGS := $(strip $(GOFLAGS) -buildvcs=false)

.PHONY: check fmt fmt-check vet test lint build run cover demo

check: fmt-check vet test lint build

fmt:
	gofmt -w cmd internal

fmt-check:
	@format_files="$$(gofmt -l cmd internal)" || exit 1; \
	if [ -n "$$format_files" ]; then \
		printf 'Нужно выполнить make fmt:\n%s\n' "$$format_files"; \
		exit 1; \
	fi

vet:
	go vet ./...

test:
	go test -race ./...

lint:
	golangci-lint run --config=.golangci.yml ./...

build:
	go build ./...

run:
	go run ./cmd/escooter-sharing

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

demo:
	go run ./cmd/escooter-sharing < docs/demo/session.txt
