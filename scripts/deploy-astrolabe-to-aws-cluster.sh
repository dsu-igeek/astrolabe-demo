#!/bin/sh
IMAGE_TAG=$1
CONF_DIR=$2
if [ -z "$IMAGE_TAG" -o -z "$CONF_DIR" ]
then
	echo "usage: deploy-astrolabe-to-cluster.sh <astrolabe image tag> <cluster dir>"
	exit 1
fi
kubectl delete --kubeconfig=/$CONF_DIR/kube/config namespace astrolabe
kubectl create --kubeconfig=/$CONF_DIR/kube/config namespace astrolabe
echo "Adding docker regcred"
./create-docker-regcred.sh $CONF_DIR
echo "Generating IVD PE config"
./gen-ivd-config.sh $CONF_DIR
echo "Generating PSQL PE config"
./gen-psql-config.sh $CONF_DIR
echo "Creating S3 config"
./gen-s3config.sh $CONF_DIR

./create-configmaps.sh $CONF_DIR
./gen-kubernetes-aws-yaml.sh $CONF_DIR $IMAGE_TAG
./install-kubernetes-aws-yaml.sh $CONF_DIR
