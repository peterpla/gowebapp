#!/usr/bin/bash

# un-comment commands that need to be performed, comment those already done

# set gcloud configuration environment for this deployment
#gcloud auth list
#gcloud config set account peterpla@gmail.com
#gcloud config list
#gcloud config set project elated-practice-224603

# create Google Cloud Storage bucket to hold encrypted resources
#gsutil mb -c standard -l $STORAGE_LOCATION -p $PROJECT_ID gs://`echo $ENCRYPTED_BUCKET`
# override the default bucket permissions
#gsutil defacl set private gs://`echo $ENCRYPTED_BUCKET`
#gsutil acl set -r private gs://`echo $ENCRYPTED_BUCKET`
# show the result
gsutil ls gs://`echo $ENCRYPTED_BUCKET`

# create Google Key Management Service keyring
#gcloud kms keyrings create $KMS_KEYRING --location $KMS_LOCATION
#gcloud kms keyrings list --location $KMS_LOCATION

# create key on that keyring
#gcloud kms keys create $KMS_KEY --location $KMS_LOCATION --keyring $KMS_KEYRING --purpose encryption
#gcloud kms keys list --location $KMS_LOCATION --keyring $KMS_KEYRING

# if NOT using GAE default service account ==> create a new service account AND adjust gsutil commands below
#gcloud iam service-accounts create lead-expert-gcs-reader
# ... with the most minimal set of permissions to access the encrypted data object
#gsutil iam ch serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com:roles/storage.objects.get \
#    gs://`echo $ENCRYPTED_BUCKET`/config.yaml.enc
#gsutil iam get gs://`echo $ENCRYPTED_BUCKET`/config.yaml.enc
# ... and the bucket
#gsutil iam ch serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com:roles/storage.objects.get \
#    gs://`echo $ENCRYPTED_BUCKET`
#gsutil iam get gs://`echo $ENCRYPTED_BUCKET`

# grant the service account access to read the App Engine deployment's env vars:
#gcloud projects add-iam-policy-binding ${PROJECT_ID} \
#  --member serviceAccount:${SA_EMAIL} \
#  --role roles/appengine.appViewer
# ... and use the KMS key
#gcloud kms keys add-iam-policy-binding ${KMS_KEY} \
#  --keyring=${KMS_KEYRING} \
#  --location=${KMS_LOCATION} \
#  --member=serviceAccount:${PROJECT_ID}@appspot.gserviceaccount.com \
#  --role=roles/cloudkms.cryptoKeyDecrypter

gcloud projects get-iam-policy ${PROJECT_ID}