#!/bin/bash

set -e
set -x

docker run -it \
    -v ${PWD}:/go/src/github.com/christopher-wong/teslatrack \
    -v ./go/src/github.com/christopher-wong/teslatrack/vendor \
    -p 8000:8000 \
    --rm \
    teslatrackd