#!/usr/bin/bash

# un-comment commands that need to be performed, comment those already done

# grant DLP User role to the App Engine application default credentials 
#gcloud projects add-iam-policy-binding $PROJECT_ID --member serviceAccount:${SA_EMAIL} --role roles/dlp.user

#gcloud projects get-iam-policy ${PROJECT_ID}