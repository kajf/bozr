#!/bin/bash

if [ $# -eq 0 ]
  then
    echo "Version argument is required."
fi

RELEASE_DIR=./release
GOARCH=amd64

mkdir -p $RELEASE_DIR

GOOS=windows
go build -o bozr.exe
zip -r $RELEASE_DIR/bozr-$1.$GOOS-$GOARCH.zip bozr.exe
rm bozr.exe

GOOS=darwin
go build -o bozr
tar -czvf $RELEASE_DIR/bozr-$1.$GOOS-$GOARCH.tar.gz bozr

GOOS=linux
go build -o bozr
tar -czvf $RELEASE_DIR/bozr-$1.$GOOS-$GOARCH.tar.gz bozr
rm ./bozr