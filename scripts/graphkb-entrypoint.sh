#!/bin/sh

go get github.com/cespare/reflex
go get github.com/go-delve/delve/cmd/dlv

reflex -r '\.go|\.yml' -s -- dlv --listen 0.0.0.0:2345 --headless=true --output=/tmp/go-graphkb --continue --accept-multiclient --api-version 2 debug cmd/go-graphkb/main.go -- --config $1 listen