.PHONY: all

all:
	docker build -t christopherwong/teslatrack -f ./Dockerfile.prod .