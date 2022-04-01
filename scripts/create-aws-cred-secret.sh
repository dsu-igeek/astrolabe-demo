#!/bin/sh
# Creates a secret from AWS environment variables
if [ -z "$AWS_ACCESS_KEY_ID" -o -z "$AWS_SECRET_ACCESS_KEY" ] 
then
	echo "AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables needed"
	exit 1
fi
kubectl create secret -n astrolabe generic aws-secret --from-literal=aws_access_key_id=$AWS_ACCESS_KEY_ID --from-literal aws_secret_access_key="$AWS_SECRET_ACCESS_KEY"
