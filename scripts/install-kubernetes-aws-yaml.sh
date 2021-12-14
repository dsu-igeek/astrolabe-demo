#!/bin/sh
cd $1
kubectl delete --kubeconfig kube/config -n astrolabe -f yaml/astrolabe-storage-class.yaml
kubectl apply --kubeconfig kube/config -n astrolabe -f yaml/astrolabe-storage-class.aws.yaml
kubectl apply --kubeconfig kube/config -n astrolabe -f yaml/astrolabe-service-account.yaml 
kubectl apply --kubeconfig kube/config -n astrolabe -f yaml/astrolabe-deployment.yaml 
