kubectl delete -n astrolabe configmap astrolabe-pes 
kubectl create -n astrolabe configmap astrolabe-pes --from-file=../../docker/astrolabe_server/conf/pes
kubectl delete -n astrolabe configmap astrolabe-conf 
kubectl create -n astrolabe configmap astrolabe-conf --from-file=../../docker/astrolabe_server/conf
