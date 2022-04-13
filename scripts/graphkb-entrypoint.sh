#!/bin/sh

go install github.com/cespare/reflex@v0.3.1
go install github.com/go-delve/delve/cmd/dlv@v1.8.2

reflex -r '\.go|\.yml' -s -- dlv --listen 0.0.0.0:2345 --headless=true --output=/tmp/go-graphkb --continue --accept-multiclient --api-version 2 debug cmd/go-graphkb/main.go -- --config $1 --log-level debug listen