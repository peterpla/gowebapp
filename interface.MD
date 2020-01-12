# Public Methods

## Typical Client Request Flow

``` text
POST /requests
        <== HTTP 202 Accepted w/ location1
GET {location1}
        <== HTTP 200 OK w/ PENDING, eta
GET {location1}
        <== HTTP 200 OK w/ COMPLETED, location2
GET {location2}
        <== HTTP 200 OK w/ transcript
```

## /requests

---

### POST /api/v1/requests

Submit a transcription Request.

TODO: authentication and authorization

#### Inputs - POST /api/v1/requests

`Content-Type` must be `application/json`.

Body, JSON: *(no key order is assumed or required)*

**Note**: `cmd/server/main.go/decodeJSONBody()` sets `http.MaxBytesReader` to `1048576` (1 mebibyte, 1024^2). If we decide to allow attached media files, we'll need to change that.

* **"customer_id"** (required) - integer

  Positive non-zero integer, seven digits. E.g., `1234567`. Primary key for identifying this customer's profile. (Allows for *10 million-1* customers.)

* **"media_uri"** (required) - string - [RFC3986](https://tools.ietf.org/html/rfc3986)

  URI to access the media file.
  
TODO: each supported external services - e.g., Twilio and Dropbox - will need an adapter to use that service's APIs to read the file.

* **"custom_config"** (optional) - structure

  **UNSUPPORTED** - future feature

  When present, the request Body includes a `config` section that overrides the customer's profile setting for transcription processing.

TODO: The customer's profile will provide many default values, e.g., destination for storing the final transcript, budgets for individual transcriptions and per-period costs (e.g., monthly maximum).

Example Request:

`POST /api/v1/requests`

Example Request Body:

```json
{
  "customer_id":   1234567,
  "media_uri": "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0"
}
```

*!!! HACK !!!*: only .MP3 files are currently accepted.

*!!! HACK !!!*: media files must already be available in Google Cloud Storage URIs in a specific GCS bucket, so the `media_uri` must be in the form `gs://`.

#### Outputs - POST /api/v1/requests

Body, JSON: *(key order is random, no ordering)*

* **"request_id"** (always) - string - [RFC4112](https://tools.ietf.org/html/rfc4122) v4

  UUID identifier of the Request, created by the server.

* **"customer_id"** (always) - integer

  From the request, unchanged

* **"media_uri"** (always) - string - [RFC3986](https://tools.ietf.org/html/rfc3986)

  From the request, unchanged.

* **"accepted_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time the Request was accepted (processing began), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"endpoint"** (on `202 Accepted`) - string

  Endpoint to `GET` for status of processing this request.

Example Response Body:

```json
{
  "request_id":   "269fd581-35aa-465d-81df-c0295034c723",
  "customer_id":  1234567,
  "media_uri":    "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0",
  "accepted_at":  "2019-12-14T16:36:47.60642Z",
  "endpoint": "/status/6697be3b-bdfa-4438-9e2a-ea1511dd0e40"
}
```

#### Response Status - POST /api/v1/requests

* 202 Accepted - success

  See `endpoint` in response for endpoint to `GET` for status of processing this request.

* 400 Bad Request

TODO: provide reason for receiving this result

* 500 Internal Server Error

---
---

## /status

---

### GET /api/v1/status/:uuid

Read the status of the previously-submitted transcription Request with `RequestID` = `uuid` ([RFC4112](https://tools.ietf.org/html/rfc4122) v4). Referred to below as *the original request*.

TODO: authentication and authorization

#### Inputs - GET /api/v1/status/:uuid

`uuid` is the UUID of the *original* request, the status of which is being checked.

`Content-Type` must be `application/json`.

Body, JSON: *(no key order is assumed or required)*

* **"customer_id"** (required) - integer

  Positive non-zero integer, seven digits. E.g., `1234567`. Primary key for identifying this customer's profile. (Allows for *10 million-1* customers.)

* **"media_uri"** (required) - string - [RFC3986](https://tools.ietf.org/html/rfc3986)

  URI to access the media file.

Both `customer_id` and `media_uri` are required for security purposes.

Example Request:

`GET /api/v1/status/:uuid`

Example Request Body:

```json
{
  "customer_id":   1234567,
  "media_uri": "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0"
}
```

#### Outputs - GET /api/v1/status

Body, JSON: *(key order is random, no ordering)*

* **"request_id"** (always) - string - [RFC4112](https://tools.ietf.org/html/rfc4122) v4

  UUID identifier of *this* Request for status; created by the server.

* **"customer_id"** (always) - integer

  From the request, unchanged.

* **"media_uri"** (always) - string - [RFC3986](https://tools.ietf.org/html/rfc3986)

  From the request, unchanged.

* **"accepted_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time *this* Request was accepted (processing began), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"completed_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time *this* Request was accepted (processing began), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"original_request_id"** (always) - string - [RFC4112](https://tools.ietf.org/html/rfc4122) v4

  UUID identifier of *the original* Request, the status of which is being reported.

* **"original_accepted_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time the *original* Request was accepted (processing began), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"original_completed_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time the *original* Request was accepted (processing began), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"original_status"** (on success) - string

  Status of *the original request*, one of `"PENDING"`, `"COMPLETED"` or `"ERROR"`:

  * `"PENDING"` : the original request is still being processed. *(See `eta` below.)*

  * `"COMPLETED"` : the original request has finished processing and the resulting transcript is available. *(See `endpoint` below.)*

  * `"ERROR"` : an error occurred during processing of the original request. *(See `original_status` below.)*

* **"eta"** (only for `status` = `"PENDING"`) - string

  Estimated duration of processing remaining. Recommended waiting at least this amount of time before the next `GET /api/v1/status` polling request. *(Experimental, this estimate may not be at all reliable.)*

* **"endpoint"** (only for `status` = `"COMPLETED"`) - string - [RFC3986](https://tools.ietf.org/html/rfc3986)

  URI to access the finished transcript.

* **"original_status"** (only for `status` = `"ERROR"`) - int - [IANA HTTP Status Code Registry](https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml)

  Error status of `4xx` or `5xx` from processing *the original* request.

Example Response Body - COMPLETED:

```json
{
  "request_id":  "269fd581-35aa-465d-81df-c0295034c723",
  "customer_id": 1234567,
  "media_uri":   "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0",
  "accepted_at": "2019-12-14T16:36:47.60642Z",
  "completed_at": "2019-12-14T16:36:47.60724Z",
  "original_status": "COMPLETED",
  "original_accepted_at": "2019-12-14T16:36:47.60642Z",
  "original_completed_at": "2019-12-14T16:36:47.60724Z",
  "endpoint":    "/transcripts/6697be3b-bdfa-4438-9e2a-ea1511dd0e40"
}
```

Example Response Body - PENDING:

```json
{
  "request_id":  "269fd581-35aa-465d-81df-c0295034c723",
  "customer_id": 1234567,
  "media_uri":   "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0",
  "accepted_at": "2019-12-14T16:36:47.60642Z",
  "original_status": "PENDING",
  "original_accepted_at": "2019-12-14T16:36:47.60642Z",
}
```

Example Response Body - ERROR:

```json
{
  "request_id":  "269fd581-35aa-465d-81df-c0295034c723",
  "customer_id": 1234567,
  "media_uri":   "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0",
  "accepted_at": "2019-12-14T16:36:47.60642Z",
  "completed_at": "2019-12-14T16:36:47.60724Z",
  "original_status": 400,
  "original_accepted_at": "2019-12-14T16:35:47.60642Z",
  "original_completed_at": "2019-12-14T16:35:47.60724Z",
}
```

#### Response Status: `GET /status`

* 200 OK - success

  This request completed successfully. See `original_status` for the status of the original request.

* 303 See Other - success

  This request completed successfully. Use `endpoint` to retrieve the finished transcript.

* 4xx Bad Request

  TODO: provide reasons for receiving 4xx results

* 5xx Internal Server Error

  TODO: provide reasons for receiving a 5xx results

---
---

## /transcripts

---

### GET /api/v1/transcripts/:uuid

Read the final transcript from the previously-submitted transcription Request with `RequestID` = `uuid` ([RFC4112](https://tools.ietf.org/html/rfc4122) v4). Referred to below as *the original request*.

TODO: authentication and authorization

#### Inputs - GET /api/v1/transcripts/:uuid

`Content-Type` must be `application/json`.

Body, JSON: *(no key order is assumed or required)*

* **"customer_id"** (required) - integer

  Positive non-zero integer, seven digits. E.g., `1234567`. Primary key for identifying this customer's profile. (Allows for *10 million-1* customers.)

* **"media_uri"** (required) - string - [RFC3986](https://tools.ietf.org/html/rfc3986)

  URI to access the media file.

Both `customer_id` and `media_uri` are required for security purposes.

Example Request:

`GET /api/v1/transcripts/:uuid`

Example Request Body:

```json
{
  "customer_id":   1234567,
  "media_uri": "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0"
}
```

#### Outputs - GET /api/v1/transcripts

Body, JSON: *(key order is random, no ordering)*

* **"request_id"** (always) - string - [RFC4112](https://tools.ietf.org/html/rfc4122) v4

  UUID identifier of *this* Request, created by the server.

* **"completed_id"** (always) - string - [RFC4112](https://tools.ietf.org/html/rfc4122) v4

  UUID identifier of *the original* Request, created by the server.

* **"customer_id"** (always) - integer

  From the request, unchanged.

* **"media_uri"** (always) - string - [RFC3986](https://tools.ietf.org/html/rfc3986)

  From the request, unchanged.

* **"accepted_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time *this* Request was accepted (processing began), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"completed_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time *this* Request was completed (processing finished), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"original_request_id"** (always) - string - [RFC4112](https://tools.ietf.org/html/rfc4122) v4

  UUID identifier of the *original* Request, the status of which is being reported.

* **"original_accepted_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time the *original* Request was accepted (processing began), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"original_completed_at"** (always) - string - [RFC3339](https://www.ietf.org/rfc/rfc3339.txt)

  Date and time the *original* Request was accepted (processing began), in RFC3339 format. Always with suffix "Z" denoting UTC/GMT (i.e., UTC/GMT offset 00:00.)

* **"transcript"** (always) - string

  Final transcript with speakers indicated and `\n` separators when speakers change.

Example Response Body:

```json
{
  "request_id": "269fd581-35aa-465d-81df-c0295034c723",
  "completed_id": "da4ae569-484d-4f59-bc52-c876058252d8",
  "customer_id": 1234567,
  "media_uri": "https://www.dropbox.com/s/0nh7urknw0fqb4h/dummy.?dl=0",
  "accepted_at": "2019-12-14T16:36:47.60642Z",
  "completed_at": "2019-12-14T16:38:12.43756Z",
  "original_accepted_at": "2019-12-14T16:35:47.60642Z",
  "original_completed_at": "2019-12-14T16:35:47.60724Z",
  "transcript": "[Speaker 1] Thank you for calling Park flooring.\n[Speaker 2] Hi, my name is Yuri.\n"
}
```

#### Response Status: `GET /api/v1/transcripts`

* 200 OK - success

* 4xx Bad Request

  TODO: provide reasons for receiving 4xx results

* 5xx Internal Server Error

  TODO: provide reasons for receiving a 5xx results

---
