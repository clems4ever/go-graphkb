#!/bin/sh

go get github.com/cespare/reflex

reflex -r '\.go|\.yml' -s -- sh -c "go run cmd/go-graphkb/main.go --config cmd/go-graphkb/config.yml listen"