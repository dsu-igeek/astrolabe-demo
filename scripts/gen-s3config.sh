#!/bin/sh
cd $1
ASTROLABE_CONFIG_DIR=astrolabe_conf
mkdir -p $ASTROLABE_CONFIG_DIR
IP=`kubectl --kubeconfig kube/config get nodes -o json | jq -r '.items[] | select(.metadata.name | test("control")|not) | .status.addresses[] | select(.type == "ExternalIP") | .address'`
S3CONFIG_FILE=$ASTROLABE_CONFIG_DIR/s3config.json
echo "{" > $S3CONFIG_FILE
echo "	\"host\":\"$IP\"," >> $S3CONFIG_FILE
echo "	\"port\":9000," >> $S3CONFIG_FILE
echo "	\"accessKey\":\"accessKey\"," >> $S3CONFIG_FILE
echo "	\"secret\":\"secretkey\"" >> $S3CONFIG_FILE
echo "}" >> $S3CONFIG_FILE

