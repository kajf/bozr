#!/bin/bash

if [ $# -eq 0 ]
  then
    echo "Version argument is required."
fi

RELEASE_DIR=./release
mkdir $RELEASE_DIR

GOOS=windows GOARCH=amd64 go build go build -o bozr.exe
zip -r $RELEASE_DIR/bozr-$1.win-x64.zip bozr.exe
rm bozr.exe

GOOS=darwin GOARCH=amd64 go build -o bozr
tar -czvf $RELEASE_DIR/bozr-$1.darwin-$GOARCH.tar.gz bozr

GOOS=linux GOARCH=amd64 go build -o bozr
tar -czvf $RELEASE_DIR/bozr-$1.linux-$GOARCH.tar.gz bozr
rm ./bozr