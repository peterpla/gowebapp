#!/usr/bin/bash

# gcloud to encrypt the configuration file:
#   https://cloud.google.com/kms/docs/quickstart
# gsutil to write encrypted configuration file to Cloud Storage:
#   https://cloud.google.com/storage/docs/quickstart-gsutil

# gcloud meta list-files-for-upload
# verbosity: debug, info, warning, error, critical, none
gcloud beta app deploy --verbosity=warning app.yaml
