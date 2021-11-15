#!/bin/sh
cd $1
ASTROLABE_CONFIG_DIR=astrolabe_conf
mkdir -p $ASTROLABE_CONFIG_DIR
IP=`kubectl --kubeconfig kube/config get nodes -o json | jq -r '.items[] | select(.metadata.name | test("control")|not) | .status.addresses[] | select(.type == "ExternalIP") | .address' | head -1`
S3CONFIG_FILE=$ASTROLABE_CONFIG_DIR/s3config.json
echo "{" > $S3CONFIG_FILE
echo "	\"host\":\"$IP\"," >> $S3CONFIG_FILE
echo "	\"port\":30900," >> $S3CONFIG_FILE
echo "	\"accessKey\":\"accesskey\"," >> $S3CONFIG_FILE
echo "	\"secret\":\"secretkey\"", >> $S3CONFIG_FILE
echo "	\"http\":true", >> $S3CONFIG_FILE
echo "	\"region\":\"astrolabe\"" >> $S3CONFIG_FILE
echo "}" >> $S3CONFIG_FILE

