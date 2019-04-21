#!/bin/bash

set -e
set -x

CGO_ENABLED=0 GOOS=linux go build -o ./cmd/teslatrackd/teslatrackd ./cmd/teslatrackd/

./cmd/teslatrackd/teslatrackd