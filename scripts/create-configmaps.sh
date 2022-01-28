#!/bin/sh
# cd $1
kubectl --kubeconfig ~/.kube/config delete -n astrolabe configmap astrolabe-pes 
kubectl --kubeconfig ~/.kube/config create -n astrolabe configmap astrolabe-pes --from-file=astrolabe_conf/pes
# Really just installs s3config
kubectl --kubeconfig ~/.kube/config delete -n astrolabe configmap astrolabe-conf 
kubectl --kubeconfig ~/.kube/config create -n astrolabe configmap astrolabe-conf --from-file=astrolabe_conf
