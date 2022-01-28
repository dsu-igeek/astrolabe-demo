#!/bin/bash

set -x
cd cmd/astrolabe_server
LD_LIBRARY_PATH=/usr/local/vmware-vix-disklib-distrib/lib64 go build -gcflags "all=-N -l"
mkdir -p ../../docker/astrolabe_server/bin
cp astrolabe_server ../../docker/astrolabe_server/bin/astrolabe_server

cd ../astrolabe-controller
make
mkdir -p ../../docker/astrolabe-controller/bin
cp bin/manager ../../docker/astrolabe_server/bin/manager

cd ../../../../minio/minio/
LD_LIBRARY_PATH=/usr/local/vmware-vix-disklib-distrib/lib64 make
mkdir -p ../../dsu-igeek/astrolabe-demo/docker/astrolabe_minio/bin
cp minio ../../dsu-igeek/astrolabe-demo/docker/astrolabe_minio/bin/minio
TAG=`date '+%b-%d-%Y-%H-%M-%S'`

cd ../../dsu-igeek/astrolabe-demo/docker/astrolabe_server
docker build -t zubron/astrolabe_server:demo .
kind load docker-image zubron/astrolabe_server:demo
# docker push zubron/astrolabe_server:$TAG

cd ../astrolabe-controller
docker build -t zubron/astrolabe-controller:demo .
kind load docker-image zubron/astrolabe-controller:demo

cd ../astrolabe_minio
docker build -t zubron/astrolabe_minio:demo .
kind load docker-image zubron/astrolabe_minio:demo
