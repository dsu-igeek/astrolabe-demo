#!/bin/sh
# usage: install-with-kustomize.sh <cluster tag> <astrolabe container tag>
BASE=`pwd`
cd $1
mkdir -p kustomize
cd kustomize
cp $BASE/../k8s/astrolabe-server/* .

