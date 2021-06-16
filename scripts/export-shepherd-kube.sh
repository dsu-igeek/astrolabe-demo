#!/bin/sh
cd $1
mkdir -p kube
cat access.json | jq -r .access | jq -r '.tkg[0].kubeconfig' > kube/config
