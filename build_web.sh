#! /bin/bash

cd src
go mod tidy
GOOS=js GOARCH=wasm go build -tags=js,wasm -o ../dist/app.wasm .