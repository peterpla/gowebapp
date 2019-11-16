#!/usr/bin/bash

# un-comment commands that need to be performed, comment those already done

# create Google Cloud Storage bucket to serve static files
#gsutil mb -c standard -l $STORAGE_LOCATION -p $PROJECT_ID gs://`echo $STATIC_FILES_BUCKET`.appspot.com
# grant read access to items in the bucket
#gsutil defacl set public-read gs://`echo $STATIC_FILES_BUCKET`.appspot.com/static

# show the result
gsutil ls gs://`echo $STATIC_FILES_BUCKET`.appspot.com/static
