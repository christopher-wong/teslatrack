FROM golang:1.11

# install dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# set working directory
RUN mkdir -p /go/src/github.com/christopher-wong/teslatrack
WORKDIR /go/src/github.com/christopher-wong/teslatrack

# add deps
ADD Gopkg.toml Gopkg.toml
ADD Gopkg.lock Gopkg.lock

# install packages, cached by docker
RUN dep ensure --vendor-only -v

EXPOSE 8000

# build binary and run bunary
CMD ["./scripts/go-build-run.sh"]

