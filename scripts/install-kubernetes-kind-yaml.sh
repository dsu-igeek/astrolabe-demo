#!/bin/bash

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $SCRIPT_DIR/kind
kubectl apply  -n astrolabe -f yaml/astrolabe-storage-class.yaml
kubectl apply  -n astrolabe -f yaml/astrolabe-service-account.yaml
kubectl apply  -n astrolabe -f yaml/astrolabe-deployment.yaml
