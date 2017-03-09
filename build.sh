#!/bin/bash

rm -rf release

mkdir release

echo "Compliling server..."
cd src && go build -o ../release/server
echo "Done server ./release/server"

echo "Compliling indexer..."
cd indexer && go build -o ../../release/indexer
echo "Done server ./release/indexer"
