.PHONY: generate all

all: build test

deps:
	go get github.com/trapped/gengen
	go get ./...

generate:
	go generate smtp/*.go
	go generate pop3/*.go
	go fmt smtp/process.go
	go fmt pop3/process.go

build: generate
	go build

test: generate
	gucumber
