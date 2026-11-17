# Changelog

[Русский](https://github.com/rusender/rusender-go/blob/main/CHANGELOG.md) · **English**

## 0.1.0

First release.

- `rusender.NewClient` for the public API: `Emails`, `Contacts`, `Lists` (`Contacts`, `Variables`), `Campaigns`,
  `Folders`, `Templates`, `Senders`, `Domains`, `SendingKeys`, `Segments`, `Webhooks`, `Files`, `Me`.
- `context.Context` as the first argument of every method; client and request settings as functional options.
- `*rusender.Error` errors with `Code`, `Details`, `RequestID`; the class of the HTTP status via `errors.Is`
  (`ErrValidation`, `ErrNotFound`, `ErrRateLimit` and others); `*NetworkError`, `ErrTimeout`.
- Retries on 502–504 and network errors; automatic `Idempotency-Key`. 429 is not retried — `ErrRateLimit`
  with `RetryAfter`.
- `Files.Upload` for campaign attachments.
- `WithResponseInto` on any method: `RequestID`, rate limit, `*http.Response` and the raw JSON of the response.
- A response outside the API contract (not the JSON envelope, a list without `items`) is
  `ErrUnexpectedResponse`, not empty data.
- Fields the API may return as `null` are pointers; moments in time are `time.Time`.
- `Nullable` for fields where `null` clears the value.
- Optional parameter fields are plain values; pointers only where an empty value is meaningful.
- Whitespace, a trailing newline and quotes around the API key are trimmed.
- Standard library only, Go 1.21+.
