#!/usr/bin/bash

# move to project root
cd /Users/peterplamondon/go/src/github.com/peterpla/lead-expert/
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

# deploy default service first, to upload source files
gcloud app deploy --verbosity=warning ./cmd/server/app.yaml

# deploy other services in parallel
gcloud app deploy --verbosity=warning --quiet ./cmd/initialRequest/app.yaml &
gcloud app deploy --verbosity=warning --quiet ./cmd/serviceDispatch/app.yaml &
gcloud app deploy --verbosity=warning --quiet ./cmd/transcriptionGCP/app.yaml &
gcloud app deploy --verbosity=warning --quiet ./cmd/transcriptionComplete/app.yaml &
wait
gcloud app deploy --verbosity=warning --quiet ./cmd/transcriptQA/app.yaml &
gcloud app deploy --verbosity=warning --quiet ./cmd/transcriptQAComplete/app.yaml &
gcloud app deploy --verbosity=warning --quiet ./cmd/tagging/app.yaml &
gcloud app deploy --verbosity=warning --quiet ./cmd/taggingComplete/app.yaml &
wait
gcloud app deploy --verbosity=warning --quiet ./cmd/taggingQA/app.yaml &
gcloud app deploy --verbosity=warning --quiet ./cmd/taggingQAComplete/app.yaml &
gcloud app deploy --verbosity=warning --quiet ./cmd/completionProcessing/app.yaml &
wait

# list all services in the current project
gcloud app services list
