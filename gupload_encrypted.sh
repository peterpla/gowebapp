#!/usr/bin/bash

gcloud kms encrypt \
    --location=$KMS_LOCATION  \
    --keyring=$KMS_KEYRING \
    --key=$KMS_KEY \
    --plaintext-file=./config.yaml \
    --ciphertext-file=./config.yaml.enc

gsutil cp ./config.yaml.enc gs://`echo $ENCRYPTED_BUCKET`

gsutil ls -l gs://`echo $ENCRYPTED_BUCKET`/config.yaml.enc

rm ./config.yaml.enc
