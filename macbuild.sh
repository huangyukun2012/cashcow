#!/bin/bash

rm -rf release

mkdir release

export GOPATH=/usr/local/lib/go:`pwd`

echo "Compliling server..."
cd src && go build -ldflags -s -o ../release/server
echo "Done server ./release/server"

echo "Compliling indexer..."
cd indexer && go build -ldflags -s -o ../../release/indexer
echo "Done server ./release/indexer"
