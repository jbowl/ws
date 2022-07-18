#!/bin/sh

echo "attempting to get cert"
aws ssm get-parameters --name jbowl.cert --with-decryption --query "Parameters[*].{Value:Value}" --region us-east-1 --output text > /app/jbowl.cert
echo "?"
/app/ws