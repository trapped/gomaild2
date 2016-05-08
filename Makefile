.PHONY: generate all

all: generate build

generate:
	go generate smtp/*.go
	go generate pop3/*.go
	go fmt smtp/process.go
	go fmt pop3/process.go

build:
	go build
