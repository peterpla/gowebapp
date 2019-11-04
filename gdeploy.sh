#!/usr/bin/bash

gcloud meta list-files-for-upload
# verbosity: debug, info, warning, error, critical, none
gcloud beta app deploy --verbosity=warning app.yaml
#    --service-account gowebapp-gcs-reader@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com 
