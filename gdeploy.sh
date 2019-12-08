#!/usr/bin/bash

# set gcloud configuration environment for this deployment
#gcloud auth list
#gcloud config set account peterpla@gmail.com
#gcloud config list
gcloud config set project `echo $PROJECT_ID`

# move to local project root
cd /Users/peterplamondon/go/src/github.com/peterpla/lead-expert/

# gcloud meta list-files-for-upload
# --verbosity= {debug, info, warning, error, critical, none}
#    --service-account lead-expert-gcs-reader@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com 

# remove old files - workaround to gcloud failing to consistently upload modified files
`echo gsutil -m rm gs://staging.$PROJECT_ID.appspot.com/**`

# deploy services
gcloud app deploy --verbosity=warning ./cmd/server/app.yaml ./cmd/initialRequest/app.yaml \
    ./cmd/serviceDispatch/app.yaml ./cmd/transcriptionGCP/app.yaml \
    ./cmd/transcriptionComplete/app.yaml ./cmd/transcriptQA/app.yaml \
    ./cmd/transcriptQAComplete/app.yaml ./cmd/tagging/app.yaml \
    ./cmd/taggingQA/app.yaml ./cmd/taggingQAComplete/app.yaml \
    ./cmd/completionProcessing/app.yaml

# list all services in the current project
gcloud app services list
