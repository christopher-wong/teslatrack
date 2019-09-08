FROM golang:1.13

# set working directory
RUN mkdir -p /go/src/github.com/christopher-wong/teslatrack
WORKDIR /go/src/github.com/christopher-wong/teslatrack

EXPOSE 8000

# build binary and run bunary
CMD ["./scripts/go-build-run.sh"]

