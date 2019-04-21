#!/bin/bash

set -e
set -x

docker build -t christopherwong/teslatrack -f Dockerfile .