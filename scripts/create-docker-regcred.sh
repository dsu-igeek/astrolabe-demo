#!/bin/sh
cd $1
DOCKER_CONFIG=~/.docker/config.json
kubectl --kubeconfig kube/config create secret -n astrolabe generic regcred \
    --from-file=.dockerconfigjson=$DOCKER_CONFIG \
    --type=kubernetes.io/dockerconfigjson
