#!/bin/sh
# Creates a secret from AWS environment variables
kubectl create secret -n astrolabe generic aws-secret --from-literal=aws_access_key_id=$AWS_ACCESS_KEY_ID --from-literal aws_secret_access_key="$AWS_SECRET_ACCESS_KEY"
