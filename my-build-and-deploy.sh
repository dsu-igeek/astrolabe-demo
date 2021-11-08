#!/bin/bash

set -x
cd cmd/astrolabe_server
LD_LIBRARY_PATH=/usr/local/vmware-vix-disklib-distrib/lib64 go build
mkdir -p ../../docker/astrolabe_server/bin
cp astrolabe_server ../../docker/astrolabe_server/bin/astrolabe_server
# cd ../astrolabe_repo_server
# LD_LIBRARY_PATH=/usr/local/vmware-vix-disklib-distrib/lib64 go build
# mkdir -p ../../docker/astrolabe_repo_server/bin
# cp astrolabe_repo_server ../../docker/astrolabe_repo_server/bin/astrolabe_repo_server
cd ../../../../minio/minio/
LD_LIBRARY_PATH=/usr/local/vmware-vix-disklib-distrib/lib64 make
mkdir -p ../../dsu-igeek/astrolabe-demo/docker/astrolabe_minio/bin
cp minio ../../dsu-igeek/astrolabe-demo/docker/astrolabe_minio/bin/minio

TAG=`date '+%b-%d-%Y-%H-%M-%S'`

cd ../../dsu-igeek/astrolabe-demo/docker/astrolabe_server
docker build --no-cache -t zubron/astrolabe_server:demo .
kind load docker-image zubron/astrolabe_server:demo
# docker push zubron/astrolabe_server:$TAG

cd ../astrolabe_minio
docker build -t zubron/astrolabe_minio:demo .
kind load docker-image zubron/astrolabe_minio:demo
# docker push dsmithuchida/astrolabe_minio:$TAG

# cd ../astrolabe_repo_server
# docker build -t zubron/astrolabe_repo_server:demo .
# kind load docker-image zubron/astrolabe_repo_server:demo
