.PHONY: build run all

all: build run

build:
	./scripts/docker-build.sh

run:
	./scripts/docker-run.sh

build-prod:
	docker build -t christopherwong/teslatrack -f ./Dockerfile.prod .