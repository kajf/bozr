#!/bin/bash

GOARCH=amd64

GOOS=windows
go build -o bozr.exe
zip -r bozr-$1.win64.zip bozr.exe

GOOS=darwin
go build -o bozr
zip -r bozr-$1.darwin.zip bozr

GOOS=linux
go build -o bozr
zip -r bozr-$1.linux.zip bozr