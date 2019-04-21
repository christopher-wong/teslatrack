.PHONY: build run all

all: build run

build:
	./scripts/docker-build.sh

run:
	./scripts/docker-run.sh