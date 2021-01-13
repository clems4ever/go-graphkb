#!/bin/bash

docker build -t go-graphkb-build .

cid=`docker create go-graphkb-build`
docker cp $cid:/node/src/go-graphkb .
docker cp $cid:/node/src/datasource-csv .
docker cp $cid:/node/src/build web/
docker rm $cid