#!/bin/sh

go vet
docker run --rm -v "$PWD":/go/src/github.com/mjwwit/traefik-transip-dns -w /go/src/github.com/mjwwit/traefik-transip-dns golang:1.11-alpine go build -v