#!/bin/bash

if [ $# -eq 0 ]
  then
    echo "Version argument is required."
fi

RELEASE_DIR=./release
export GOARCH=amd64

mkdir -p $RELEASE_DIR

MD5_SUM=""
if [ -x "$(command -v md5sum)" ]; then
  MD5_SUM=md5sum
fi

if [ -x "$(command -v md5)" ]; then
  MD5_SUM=md5
fi

# Windows build
export GOOS=windows
go build -o bozr.exe

if [ -n $MD5_SUM ]; then
  echo "Windows: " "$($MD5_SUM ./bozr.exe)"
fi

zip -r $RELEASE_DIR/bozr-$1.$GOOS-$GOARCH.zip bozr.exe
rm bozr.exe

echo "------------------------------"

# MacOS build
export GOOS=darwin
go build -o bozr

if [ -n $MD5_SUM ]; then
  echo "Darwin: " "$($MD5_SUM ./bozr)"
fi

tar -czvf $RELEASE_DIR/bozr-$1.$GOOS-$GOARCH.tar.gz bozr

echo "------------------------------"

# Linux build
export GOOS=linux
go build -o bozr

if [ -n $MD5_SUM ]; then
  echo "Linux: " "$($MD5_SUM ./bozr)"
fi

tar -czvf $RELEASE_DIR/bozr-$1.$GOOS-$GOARCH.tar.gz bozr
rm ./bozr