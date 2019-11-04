# gowebapp

A generic Go web app with APIs and html/template UI, intended to be a "Go app starter kit", i.e., a decent starting point for something real.

## Command Line Interface

- `--help` to display usage info
- `--port=80` to listen on port `:80` (default is `:8080`)
- `--v` to enable verbose output

## Environment Variables

Used to access configuration file on Google Cloud Storage

- `ENCRYPTED_BUCKET` string, name of GCS bucket that contains the configuration file, must be unique within GCS
- `STORAGE_LOCATION` string, GCP deployment region of storage bucket
- `CONFIG_FILE` string, filename of config file

Used to decrypt the configuration file

- `PROJECT_ID` string, GCP project ID
- `KMS_LOCATION` string, GCP deployment region of keyring
- `KMS_KEYRING` string, name of KMS keyring
- `KMS_KEY` string, name of KMS key on that keyring

Google App Engine requires checking `PORT` before calling `ListenAndServe()`

- `PORT` int, port number to listen on

## Config File

- `AppName` string, user-friendly name of application
- `Version` string, x.y.z following Semantic Versioning (SemVer)
- `Description` string, user-friendly description of application
