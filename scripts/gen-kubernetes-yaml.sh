#!/bin/sh
# usage: install-with-kustomize.sh <cluster tag> <astrolabe container tag>
BASE=`pwd`
cd $1
mkdir -p yaml
sed -e "s/latest/$2/" < $BASE/../k8s/astrolabe-server/astrolabe-deployment.yaml > ./yaml/astrolabe-deployment.yaml
cp $BASE/../k8s/astrolabe-server/astrolabe-storage-class.yaml ./yaml
cp $BASE/../k8s/astrolabe-server/astrolabe-service-account.yaml ./yaml
cp $BASE/../k8s/astrolabe-server/local-path-storage-class.yaml ./yaml
