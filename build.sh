#!/usr/bin/env bash

rm -rf ./bin/*

# linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags "-extldflags -static" -o ./webapp-go ./main.go

# windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build --ldflags "-extldflags -static" -o ./webapp-go.exe ./main.go
