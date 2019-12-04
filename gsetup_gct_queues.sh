#!/usr/bin/bash

# un-comment commands you need to perform, comment those already done
# see "Cloud Task queues with App Engine targets"
# <https://cloud.google.com/tasks/docs/dual-overview#appe>

# List all available queues
#gcloud tasks queues list

# 2. "Create a queue"
# <https://cloud.google.com/tasks/docs/creating-queues>
#gcloud tasks queues create InitialRequest
#gcloud tasks queues create ServiceDispatch
#gcloud tasks queues create TranscriptionGCP
#gcloud tasks queues create TranscriptionComplete
#gcloud tasks queues create TranscriptQA
#gcloud tasks queues create TranscriptQAComplete
#gcloud tasks queues create Tagging
#gcloud tasks queues create TaggingComplete
#gcloud tasks queues create TaggingQA
#gcloud tasks queues create TaggingQAComplete
#gcloud tasks queues create CompletionProcessing
#gcloud tasks queues create Persistence
#gcloud tasks queues create PersistenceComplete
#gcloud tasks queues create Reporting
#gcloud tasks queues create ReportingComplete
#gcloud tasks queues create Billing
#gcloud tasks queues create BillingComplete

# "It can take a few minutes for a newly created queue to be available."
# TODO: sleep or loop until queue listed?

# TODO: adjust queue properties like rateLimit, retryConfig

# Use Stackdriver logging with Cloud Tasks queues.
# The log-sampling-ratio value indicates what percentage of the
# operations on the queue are logged. Turn off logging by setting the
# flag to 0.0.
#gcloud beta tasks queues update InitialRequest --log-sampling-ratio=1.0

# "use describe to verify that your queue was created successfully"
#gcloud tasks queues describe InitialRequest > gcp_logs/queueDetails_InitialRequest.txt
#gcloud tasks queues describe ServiceDispatch > gcp_logs/queueDetails_ServiceDispatch.txt
#gcloud tasks queues describe TranscriptionGCP > gcp_logs/queueDetails_TranscriptionGCP.txt
#gcloud tasks queues describe TranscriptionComplete > gcp_logs/queueDetails_TranscriptionComplete.txt
#gcloud tasks queues describe TranscriptQA > gcp_logs/queueDetails_TranscriptQA.txt
#gcloud tasks queues describe TranscriptQAComplete > gcp_logs/queueDetails_TranscriptQAComplete.txt
#gcloud tasks queues describe Tagging > gcp_logs/queueDetails_Tagging.txt
#gcloud tasks queues describe TaggingComplete > gcp_logs/queueDetails_TaggingComplete.txt
#gcloud tasks queues describe TaggingQA > gcp_logs/queueDetails_TaggingQA.txt
#gcloud tasks queues describe TaggingQAComplete > gcp_logs/queueDetails_TaggingQAComplete.txt
#gcloud tasks queues describe CompletionProcessing > gcp_logs/queueDetails_CompletionProcessing.txt
#gcloud tasks queues describe Persistence > gcp_logs/queueDetails_Persistence.txt
#gcloud tasks queues describe PersistenceComplete > gcp_logs/queueDetails_PersistenceComplete.txt
#gcloud tasks queues describe Reporting > gcp_logs/queueDetails_Reporting.txt
#gcloud tasks queues describe ReportingComplete > gcp_logs/queueDetails_ReportingComplete.txt
#gcloud tasks queues describe Billing > gcp_logs/queueDetails_Billing.txt
#gcloud tasks queues describe BillingComplete > gcp_logs/queueDetails_BillingComplete.txt

# List all available queues
gcloud tasks queues list