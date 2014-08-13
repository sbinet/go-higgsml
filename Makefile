.PHONY: all build run

all: build
	@echo "done"

build:
	@go get -v .

run: build
	@echo "=== run training ==="
	go-higgsml -train training.csv trained.dat

	@echo "=== run prediction ==="
	go-higgsml test.csv trained.dat scores_test.csv


