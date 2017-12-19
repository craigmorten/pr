SHELL := /bin/bash
TARGET := $(shell echo $${PWD\#\#*/})
VERSION := 1.0.0
BUILD := `git rev-parse HEAD`
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

.PHONY: all build clean install uninstall fmt simplify check run
all: check deps build

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o $(TARGET)

build: $(TARGET)
	@true

clean:
	@rm -f $(TARGET)

deps:
	@glide install --strip-vendor

test:
	@go test $(go list ./... | grep -v /vendor/) -coverprofile coverage.out -covermode count
	@go tool cover -func=coverage.out

coverage:
	@go tool cover -html=coverage.out -o coverage.html

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)

check:
	@test -z $(shell gofmt -l main.go | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
	@go tool vet ${SRC}

run: deps
	@$(TARGET)