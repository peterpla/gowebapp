# Public Methods

## /requests

---

### POST /requests

Submit a transcription Request.

TODO: authentication and authorization

#### Inputs - POST /requests

Body, JSON: *(no key order is assumed or required)*

* **"CustomerID"** (required) - string

    Primary key for identifying this customer's profile.

* **"MediaFileURL"** (required) - string

    URL to access the media file.
  
TODO: each supported external services - e.g., Twilio and Dropbox - will need an adapter to use that service's APIs to read the file.

* **"CustomConfig"** (optional) - boolean

    Flag that indicates whether the request Body includes a `config` section that overrides the customer's profile setting for transcription processing.

TODO: The customer's profile will provide many default values, e.g., destination for storing the final transcript, budgets for individual transcriptions and per-period costs (e.g., monthly maximum).

Example Request:

`POST /requests`

Example Request Body:

```json
{
  "CustomerID":   "0123456",
  "MediaFileURI": "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0"
}
```

#### Outputs - POST /requests

Body, JSON: *(key order is random, no ordering)*

* **"RequestID"** (always) - string

    UUID identifier of the Request, created by the server. [RFC 4112](https://tools.ietf.org/html/rfc4122) version 4.

* **"CreatedAt"** (always) - string - RFC3339

    Date and time the Request was created, in RFC3339 format. Always expressed as UTC/GMT.

TODO: provide estimated cost feedback to the requestor.

* **"EstimatedCost"** (always) - float

    Server-generated estimate of the cost (USD) to produce the transcription, based on the configuration provided (if `CustomConfig` was specified) or the configuration from the customer's profile.

Example Response Body:

```json
{
  "RequestID":      "269fd581-35aa-465d-81df-c0295034c723",
  "CreatedAt":      "2019-06-22T21:54:56.714Z",
}
```

#### Response Status

* 201 Created - success

TODO: consider 202 Accepted, see [REST and long-running jobs](https://farazdagi.com/2014/rest-and-long-running-jobs/) and [Long running REST API with queues](https://stackoverflow.com/a/33011965/10649045) for detailed discussions of how to handle this.

* 400 Bad Request

TODO: provide reason for receiving this result

* 500 Internal Server Error

---

### GET /requests/\{RequestID}

Read the transcription Request with the specified `RequestID`.

---

### GET /requests

Read the first 20 Requests. `GET /requests` is equivalent to `GET /requests?limit=20&offset=0`, see below.

---

### GET /requests?limit=\{n}&offset=\{m}

Read `n` Requests starting at offset `m`. Records will be returned "most recent first" order (i.e., in reverse chronological order) based on the requests' `CreatedAt` values.

---

### PATCH /requests/\{RequestID}

Update the transcription Request with the specified `RequestID`, using the provided key/value pair(s). The updated Request is returned as JSON in the response body.

---

### DELETE /requests/\{RequestID}

Cancel the transcription Request with the specified `RequestID`.

JSON in the response body includes any costs already incurred.

---
---
