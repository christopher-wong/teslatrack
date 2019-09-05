.PHONY: all build-prod

all:
	docker-compose up

build-prod:
	docker build -t christopherwong/teslatrack:0.0.8 -f ./Dockerfile.prod .