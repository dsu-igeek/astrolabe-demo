#!/bin/sh
cd $1
VC=`cat access.json | jq -r .access | jq -r .vc[0].ip`
USER=`cat access.json | jq -r .access | jq -r .vc[0].vimUsername`
PASSWORD=`cat access.json | jq -r .access | jq -r .vc[0].vimPassword`
DC=`cat access.json | jq -r .access | jq -r .cluster[0].datacenter`
PES_DIR=astrolabe_conf/pes
mkdir -p $PES_DIR
IVD_FILE=$PES_DIR/ivd.pe.json
echo "{" > $IVD_FILE
echo "	\"VirtualCenter\":\"$VC\"," >> $IVD_FILE
echo "	\"insecureVC\":\"Y\"," >> $IVD_FILE
echo "	\"user\":\"$USER\"," >> $IVD_FILE
echo "	\"password\":\"$PASSWORD\"," >> $IVD_FILE
echo "	\"port\":\"443\"," >> $IVD_FILE
echo "	\"insecure-flag\":\"true\"," >> $IVD_FILE
echo "	\"cluster-id\":\"123\"," >> $IVD_FILE
echo "	\"datacenters\":\"$DC\"" >> $IVD_FILE
echo "}" >> $IVD_FILE

