# Public Methods

## /requests

---

### POST /requests

Submit a transcription Request.

TODO: authentication and authorization

#### Inputs - POST /requests

Body, JSON: *(no key order is assumed or required)*

* **"customer_id"** (required) - string

    Primary key for identifying this customer's profile.

* **"media_uri"** (required) - string

    URI to access the media file.
  
TODO: each supported external services - e.g., Twilio and Dropbox - will need an adapter to use that service's APIs to read the file.

* **"custom_config"** (optional) - boolean

    Flag that indicates whether the request Body includes a `config` section that overrides the customer's profile setting for transcription processing.

TODO: The customer's profile will provide many default values, e.g., destination for storing the final transcript, budgets for individual transcriptions and per-period costs (e.g., monthly maximum).

Example Request:

`POST /requests`

Example Request Body:

```json
{
  "customer_id":   "0123456",
  "media_uri": "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0"
}
```

**HACK!!!**: only Google Cloud Storage URIs in the form `gs://` are currently accepted, and the media file to be transcribed must already be present in the GCS bucket `elated-practice-224603.appspot.com/audio_uploads/`.

#### Outputs - POST /requests

Body, JSON: *(key order is random, no ordering)*

* **"request_id"** (always) - string - RFC4112

    UUID identifier of the Request, created by the server. [RFC4112](https://tools.ietf.org/html/rfc4122) version 4.

* **"customer_id"** (always) - string

    From the request, unchanged

* **"media_uri"** (always) - string

    From the request, unchanged.

* **"accepted_at"** (always) - string - RFC3339

    Date and time the Request was accepted (processing began), in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

Example Response Body:

```json
{
  "request_id":   "269fd581-35aa-465d-81df-c0295034c723",
  "customer_id":  "0123456",
  "media_uri":    "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0",
  "accepted_at":  "2019-12-14T16:36:47.60642Z",
}
```

#### Response Status

* 202 Accepted - success

TODO: see [REST and long-running jobs](https://farazdagi.com/2014/rest-and-long-running-jobs/) and [Long running REST API with queues](https://stackoverflow.com/a/33011965/10649045) for detailed discussions of how to return status to the client upon completion of a long-running request.

* 400 Bad Request

TODO: provide reason for receiving this result

* 500 Internal Server Error

---

### GET /requests/\{RequestID}

Read the transcription Request with the specified `RequestID`.

---

### GET /requests

Read the first 20 Requests, in "most recent first" order (i.e., reverse chronological order) based on the requests' `AcceptedAt` values.

`GET /requests` is equivalent to `GET /requests?limit=20&offset=0`, see below.

---

### GET /requests?limit=\{n}&offset=\{m}

Read `n` Requests starting at offset `m`. Records will be returned "most recent first" order (i.e., reverse chronological order) based on the requests' `AcceptedAt` values.

---

### PATCH /requests/\{RequestID}

Update the transcription Request with the specified `RequestID`, using the provided key/value pair(s). The updated Request is returned as JSON in the response body.

---

### DELETE /requests/\{RequestID}

Cancel the transcription Request with the specified `RequestID`.

JSON in the response body includes any costs already incurred.

---
---
