NAME = $(notdir $(PWD))
VERSION = $(shell printf "%s.%s" $$(git rev-list --count HEAD) $$(git rev-parse --short HEAD))
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

test: 
	@echo :: run tests
	go test -v -race -cover -covermode=atomic -coverprofile=coverage.txt ./...

build:  $(OUTPUT)
	CGO_ENABLED=0 GOOS=linux go build -o bin/app \
		-ldflags "-X main.version=$(VERSION)" \
		-gcflags "-trimpath $(GOPATH)/src"
	mkdir -p azure/bin && cp bin/app azure/bin/app

$(OUTPUT):
	mkdir -p $(OUTPUT)
