#!/bin/sh
GITSHA=`git log -1 --pretty=format:%h`
DATE=`date +%Y-%m-%d-%H-%M`
TAG="$GITSHA"-"$DATE"
echo $TAG
TAG=$TAG make containers

cd ../../dsu-igeek/astrolabe-demo/docker/astrolabe_server
docker build -t dsmithuchida/astrolabe_server:$TAG .
docker push dsmithuchida/astrolabe_server:$TAG

cd ../astrolabe_minio
docker build -t dsmithuchida/astrolabe_minio:$TAG .
docker push dsmithuchida/astrolabe_minio:$TAG

cd ../astrolabe_repo_server
docker build -t dsmithuchida/astrolabe_repo_server:$TAG .
docker push dsmithuchida/astrolabe_repo_server:$TAG
