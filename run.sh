#! /bin/bash

cd src
go mod tidy
go build -o ../dist/app .
cd ..
./dist/app