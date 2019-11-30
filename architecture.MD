# Architecture - XYZZY Service

## An HTTP API for automated sales call transcription

## Language

- **Actionable Sales Information**: data extracted during transcription processing that enables the customer to pursue sales opportunities
- **Billing**: record billable activities and periodically charge the customer, e.g., for transcriptions delivered and storage used
- **Customer**
- **External service**: e.g., Sonix.ai, Google Speech-to-Text, Dropbox
- **Media**: the audio or video file to be transcribed
- **Persistence**: storing work products for later use by the customer or us
- **Prep**: perform setup steps required for the next stage of request processing
- **Queue**: an ordered sequence of requests at various stages of processing
- **Reporting**: periodically providing information the customer wants, delivered as they prefer
- **Request**: a customer-initiated event to begin producing a transcription; may be initiated by an external service on behalf of the customer, e.g., a new Twilio call recording is available
- **Review**: a human corrects/enhances the transcript produced by an external service, improving the transcript to be more accurate and valuable
- **Reviewer**: a human that performs reviews
- **Transcript**: a text file with the written transcript produced from the provided media file
- **Validation**: confirm a request meets the requirements to proceed to the next stage of request processing

## Services

**Request**: create, enqueue, dequeue, process, complete ... a request
**Transcript**: create, review, tag, complete, deliver ... a transcript
**Reporting**: create, complete, deliver ... a report

## Value Objects

Customer, request, media file, transcript, tagging metadata

## Events

**Request**: rejected, created, added to queue, dispatched to worker, process step completed and removed from queue, retried, special handling required
**Transcript**: created, under review, review completed, transcripion completed, delivered to customer, special handling required
**Customer**: added, modified, transcription requested, transcript delivered, report delivered, service charge allocation added, billing event, external service resource billed
**External Services**: processing by service requested, processing completed, service resource consumed (capturing key metrics), service charge received, service charges matched to customer usage, service charges allocated to customers for billing
**Billing**: payment method established, payment method verified, external service consumed (e.g., seconds of processing), periodic charges calculated (e.g., GB of storage), customer payment initiated (e.g., credit card charged), customer payment received

## Repository

**Customer Repository**: enables the customer to review previously completed recordings, transcripts and reports. Contains: original audio file, delivered transcript, reports

**Transcription Artifacts**: enables us to revisit previously completed requests and work product to develop customized ML models, assess improvement by external services. Contains: delivered transcript, tagging metadata; intermediate transcripts; all associated with customer's original audio file

## Packages

``` text
cmd (things you start)
|__data
|  |__main.go (add sample data, etc.)
|__server
   |__main.go (HTTP server, routing, etc.)
pkg
|__adding (enqueue)
|__completing
|__delivering
|__listing
|__removing (dequeue)
|__reviewing
|__storage
|__submitting
|__[value objects - `struct`s]
```

## GAE Services

Google App Engine services as defined in `app.yaml` files.

``` text
default (./cmd/server, app startup)
|__initial-request (./cmd/initial-request, incoming requests for transcription)
```

## Cloud Tasks Queues

Google Cloud Tasks queues and flow: *(see Airtable, "Flooring: Transcription Pipeline" for details)*

1. **InitialRequest**: tasks added by `default` service, handled by `initial-request` service, `/task_handler` endpoint, implemented in `./cmd/initialRequest/main.go`
1. **ServiceDispatch**: tasks added by `initial-request` service, handled by `service-dispatch` service, `/task_handler` endpoint, implemented in `./cmd/serviceDispatch/main.go`
1. **TranscriptionGCP**: tasks added by `service-dispatch` service, handled by `transcription-GCP` service, `/task_handler` endpoint, implemented in `./cmd/transcriptionGCP/main.go`
1. **TranscriptionComplete**: tasks added by `transcription-GCP` service, handled by `transcription-complete` service, `/task_handler` endpoint, implemented in `./cmd/transcriptionComplete/main.go`
1. **TranscriptQA**: tasks added by `transcription-complete` service, handled by `transcript-QA` service, `/task_handler` endpoint, implemented in `./cmd/transcriptQA/main.go`
1. **TranscriptQAComplete**: tasks added by `transcript-QA` service, handled by `transcript-QA-complete` service, `/task_handler` endpoint, implemented in `./cmd/transcriptionQAComplete/main.go`
1. **Tagging**: tasks added by `transcript-QA-complete` service, handled by `tagging` service, `/task_handler` endpoint, implemented in `./cmd/tagging/main.go`
1. **TaggingComplete**: tasks added by `tagging` service, handled by `tagging-complete` service, `/task_handler` endpoint, implemented in `./cmd/taggingComplete/main.go`
1. **TaggingQA**: tasks added by `tagging-complete` service, handled by `tagging-QA` service, `/task_handler` endpoint, implemented in `./cmd/taggingQA/main.go`
1. **TaggingQAComplete**: tasks added by `tagging-QA` service, handled by `transcript-request-complete` service, `/task_handler` endpoint, implemented in `./cmd/transcriptionRequestComplete/main.go`
