#!/bin/bash

GOARCH=amd64

GOOS=windows
go build -o bozr.exe
zip -r bozr-$1.win64.zip bozr.exe

GOOS=darwin
go build -o bozr
tar -czvf bozr-$1.darwin.tar.gz bozr

GOOS=linux
go build -o bozr
tar -czvf bozr-$1.linux.tar.gz bozr