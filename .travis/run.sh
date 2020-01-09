#!/bin/bash

go test -v ./...

pushd web && npm ci && npm run build && popd
go build -o go-graphkb cmd/go-graphkb/main.go
go build -o importer-csv cmd/importer-csv/main.go

