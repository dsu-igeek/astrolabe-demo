#!/bin/sh
kubectl create -n astrolabe secret generic regcred \
    --from-file=.dockerconfigjson=/home/dsmithuchida/.docker/config.json \
    --type=kubernetes.io/dockerconfigjson
