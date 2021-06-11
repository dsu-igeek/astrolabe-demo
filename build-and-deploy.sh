#!/bin/sh

cd cmd/astrolabe_server
go build
cp astrolabe_server ../../docker/astrolabe_server/bin/astrolabe_server
cd ../../../../minio/minio/
make
cp minio ../../dsu-igeek/astrolabe-demo/docker/astrolabe_minio/bin/minio

TAG=`date '+%b-%d-%Y-%H-%M-%S'`

cd ../../dsu-igeek/astrolabe-demo/docker/astrolabe_server
docker build -t dsmithuchida/astrolabe_server:$TAG .
docker push dsmithuchida/astrolabe_server:$TAG

cd ../astrolabe_minio
docker build -t dsmithuchida/astrolabe_minio:$TAG .
docker push dsmithuchida/astrolabe_minio:$TAG

