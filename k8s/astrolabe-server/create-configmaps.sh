kubectl create -n astrolabe configmap astrolabe-pes --from-file=conf/pes
kubectl create -n astrolabe configmap astrolabe-s3config --from-file=conf/s3config.json
