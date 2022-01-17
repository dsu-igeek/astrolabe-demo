#!/bin/sh

REGISTRY="${REGISTRY:-dsmithuchida}"

echo "Building controller"
cd cmd/astrolabe-controller
make generate
make manifests
make
mkdir -p ../../docker/astrolabe-controller/bin
cp bin/manager ../../docker/astrolabe-controller/bin/manager

echo "Building server"
cd ../astrolabe_server
LD_LIBRARY_PATH=/usr/local/vmware-vix-disklib-distrib/lib64 go build
mkdir -p ../../docker/astrolabe_server/bin
cp astrolabe_server ../../docker/astrolabe_server/bin/astrolabe_server

echo "Building repo server"
cd ../astrolabe_repo_server
LD_LIBRARY_PATH=/usr/local/vmware-vix-disklib-distrib/lib64 go build
mkdir -p ../../dsu-igeek/astrolabe-demo/docker/astrolabe_repo_server/bin
cp astrolabe_repo_server ../../dsu-igeek/astrolabe-demo/docker/astrolabe_repo_server/bin/astrolabe_repo_server

echo "Building minio"
cd ../../../../minio/minio/
LD_LIBRARY_PATH=/usr/local/vmware-vix-disklib-distrib/lib64 make
mkdir -p ../../dsu-igeek/astrolabe-demo/docker/astrolabe_minio/bin
cp minio ../../dsu-igeek/astrolabe-demo/docker/astrolabe_minio/bin/minio
TAG=`date '+%b-%d-%Y-%H-%M-%S'`

cd ../../dsu-igeek/astrolabe-demo/docker/astrolabe-controller
docker build -t $REGISTRY/astrolabe-controller:$TAG .
docker push $REGISTRY/astrolabe-controller:$TAG

cd ../../dsu-igeek/astrolabe-demo/docker/astrolabe_server
docker build -t $REGISTRY/astrolabe_server:$TAG .
docker push $REGISTRY/astrolabe_server:$TAG

cd ../astrolabe_minio
docker build -t $REGISTRY/astrolabe_minio:$TAG .
docker push $REGISTRY/astrolabe_minio:$TAG

cd ../astrolabe_repo_server
docker build -t $REGISTRY/astrolabe_repo_server:$TAG .
docker push $REGISTRY/astrolabe_repo_server:$TAG
