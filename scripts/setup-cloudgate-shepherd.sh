#!/bin/bash

set -eu

read -p "account_group name: " ACCOUNT_GROUP_NAME
read -p "Cloudgate API Client ID: " CLOUDGATE_ID
read -s -p "Cloudgate API Client Secret: " CLOUDGATE_SECRET
echo
read -p "Organizational Unit ID: " OU_ID
read -p "Account ID: " ORG_ACCOUNT_ID
read -p "Master Account ID: " MASTER_ACCOUNT_ID
read -p "AWS Region: " AWS_REGION

# Create Account Group
sheepctl account-group create \
  --name ${ACCOUNT_GROUP_NAME} \
  --kind aws-cloudgate

# Create Account:
sheepctl account create \
  --group "$ACCOUNT_GROUP_NAME" \
  --name "org_account_id-${ORG_ACCOUNT_ID}" \
  --data "$( \
    jq -n \
      --arg 'cloudgate_id' "$CLOUDGATE_ID" \
      --arg 'cloudgate_token' "$CLOUDGATE_SECRET" \
      --arg 'org_unit_id' "$OU_ID" \
      --arg 'org_account_id' "$ORG_ACCOUNT_ID" \
      --arg 'master_account_id' "$MASTER_ACCOUNT_ID" \
      --arg 'aws_region' "$AWS_REGION" \
       '{
        "kind": "aws-cloudgate",
        "version": "1.0.0",
        "cloudgate_credentials": {
            "id": $cloudgate_id,
            "token": $cloudgate_token,
        },
        "cloudgate_account": {
            "master_account_id": $master_account_id,
            "org_unit_id": $org_unit_id,
            "org_account_id": $org_account_id,
            "default_region": $aws_region,
        },
        "extra_data": {}
    }'
  )"\
  --maximum 1

