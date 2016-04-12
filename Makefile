.PHONY: generate all

all: generate build

generate:
	go generate smtp/*.go
	go fmt smtp/process.go

build:
	go build