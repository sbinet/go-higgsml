.PHONY: all build run

all: build
	@echo "done"

build:
	@go get -v .

run: build
	go-higgsml

