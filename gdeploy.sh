#!/usr/bin/bash

# move to project root
cd /Users/peterplamondon/go/src/github.com/peterpla/lead-expert/

# upload static files to the bucket
#gsutil -m rsync -R ./public gs://`echo $STATIC_FILES_BUCKET`.appspot.com/static

# gcloud meta list-files-for-upload
# --verbosity= {debug, info, warning, error, critical, none}
#    --service-account lead-expert-gcs-reader@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com 

# remove old files - workaround to gcloud failing to upload modified files
gsutil -m rm gs://staging.elated-practice-224603.appspot.com/**

# deploy services
gcloud app deploy --verbosity=warning ./cmd/server/app.yaml ./cmd/initialRequest/app.yaml \
    ./cmd/serviceDispatch/app.yaml ./cmd/transcriptionGCP/app.yaml \
    ./cmd/transcriptionComplete/app.yaml ./cmd/transcriptQA/app.yaml \
    ./cmd/transcriptQAComplete/app.yaml ./cmd/tagging/app.yaml \
    ./cmd/taggingComplete/app.yaml ./cmd/taggingQA/app.yaml \
    ./cmd/taggingQAComplete/app.yaml ./cmd/completionProcessing/app.yaml

# list all services in the current project
gcloud app services list
