@echo off
cd src
go mod tidy
go build -o ..\dist\app.exe .
cd ..
dist\app.exe