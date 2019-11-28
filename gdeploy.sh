#!/usr/bin/bash

# move to project root
cd /Users/peterplamondon/go/src/github.com/peterpla/gowebapp/

# upload statis files to the bucket
#gsutil -m rsync -R ./public gs://`echo $STATIC_FILES_BUCKET`.appspot.com/static

# gcloud meta list-files-for-upload
# --verbosity= {debug, info, warning, error, critical, none}
#    --service-account gowebapp-gcs-reader@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com 

# deploy services
gcloud app deploy --verbosity=warning ./cmd/server/app.yaml ./cmd/wInitialRequest/app.yaml

# list all services in the current project
gcloud app services list
