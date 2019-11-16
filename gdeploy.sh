#!/usr/bin/bash

# upload statis files to the bucket
gsutil -m rsync -R ./public gs://`echo $STATIC_FILES_BUCKET`.appspot.com/static

# gcloud meta list-files-for-upload
# verbosity: debug, info, warning, error, critical, none
gcloud app deploy --verbosity=warning app.yaml
#    --service-account gowebapp-gcs-reader@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com 
