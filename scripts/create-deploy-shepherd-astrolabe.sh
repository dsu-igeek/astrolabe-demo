#!/bin/sh
IMAGE_TAG=$1
TAG=`date '+%b-%d-%Y-%H-%M-%S'`
#TAG=Jun-12-2021-11-46-01
CONF_DIR=~/astrolabe-clusters/$TAG
mkdir -p $CONF_DIR
sheepctl lock create -f shepherd-recipes/tkgm-vsphere.json -o $CONF_DIR/access.json --lifetime 3d
echo "Export kube config"
./export-shepherd-kube.sh $CONF_DIR
kubectl delete --kubeconfig=/$CONF_DIR/kube/config namespace astrolabe
kubectl create --kubeconfig=/$CONF_DIR/kube/config namespace astrolabe
echo "Adding docker regcred"
./create-docker-regcred.sh $CONF_DIR
echo "Generating IVD PE config"
./gen-ivd-config.sh $CONF_DIR
echo "Creating S3 config"

./gen-s3config.sh $CONF_DIR
./create-configmaps.sh $CONF_DIR
./gen-kubernetes-yaml.sh $CONF_DIR $IMAGE_TAG
./install-kubernetes-yaml.sh $CONF_DIR
