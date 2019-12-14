# lead-expert

An automated service to deliver tagged call transcripts from audio/video recordings

## Command Line Interface

- `--help` to display usage info
- `--v` to enable verbose output

## Environment Variables

Credentials for Google Cloud Platform

- `PROJECT_ID` string, GCP project ID *(e.g., `silly-name-123456`)*
- `PROJECT_NAME` string, should be `lead-expert`
- `SERVICE_ACCOUNT_EXTENSION` string, GCP-assigned value used in service account `.json` file name
- `GOOGLE_ACCOUNT_CREDENTIALS` string, combines path, `PROJECT_NAME` and `SERVICE_ACCOUNT_EXTENSION` to identify `.json` file
- `SA_EMAIL` string, service account email used by each service

Used to access Google Cloud Storage

- `ENCRYPTED_BUCKET` string, name of GCS bucket that contains the configuration file, must be unique within GCS
- `STORAGE_LOCATION` string, GCP deployment region of storage bucket
- `CONFIG_FILE` string, filename of config file within `ENCRYPTED_BUCKET`

Used with Google Key Management Service to decrypt the config file

- `KMS_LOCATION` string, GCP deployment region of keyring
- `KMS_KEYRING` string, name of KMS keyring
- `KMS_KEY` string, name of KMS key on that keyring

Used with Google Cloud Tasks

- `TASKS_LOCATION` string, GCP deployment region of Cloud Tasks queues

Service-specific configuration

- `TASKS_[taskname]_SERVICENAME`, string, name of Google App Engine service
- `TASKS_[taskname]_WRITE_TO_Q`, string, name of Cloud Tasks queue that is next in the processing pipeline
- `TASKS_[taskname]_SVC_TO_HANDLE_REQ`, string, name of Google App Engine service that will handle the next processing pipeline stage
- `TASKS_[taskname]_PORT`, string, *when running locally*, the service will use this port (e.g., "8080"). Ignored when running on GAE, as GAE assigns a port to each service.

Google App Engine requires checking `PORT` before calling `ListenAndServe()`

- `PORT` string, port number to listen on (e.g., "8080"). *Checked as required by GAE, but ignored*

## Config File

- `AppName` string, user-friendly name of application
- `Version` string, x.y.z following [Semantic Versioning 2.0.0](https://semver.org/) (SemVer)
- `Description` string, user-friendly description of application
