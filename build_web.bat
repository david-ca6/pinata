@echo off
cd src
go mod tidy
set GOOS=js
set GOARCH=wasm
go build -tags=js,wasm -o ..\dist\app.wasm .
cd ..
copy websrc\* dist\