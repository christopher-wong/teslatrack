.PHONY: all build-prod

all:
	docker-compose up

build:
	docker build -t registry.digitalocean.com/teslatrack/teslatrack_api:0.0.9 -f ./Dockerfile.prod .