# RuSender Go SDK

[Русский](https://github.com/rusender/rusender-go/blob/main/README.md) · **English**

The official Go library for the [RuSender public API](https://rusender.ru/docs/). Standard library only,
Go 1.21 or newer.

> Development happens in RuSender's internal repository; GitHub is a mirror. Report bugs and ideas in Issues.

## Installation

```bash
go get github.com/rusender/rusender-go
```

## Quick start

Create an API key in your account: **Integrations → API**. Pass the key to `NewClient`; with an empty string
it is read from `RUSENDER_API_KEY`. Whitespace, a trailing newline and quotes around the key are trimmed.

```go
import "github.com/rusender/rusender-go"

client, err := rusender.NewClient(os.Getenv("RUSENDER_API_KEY"))
if err != nil {
	log.Fatal(err)
}

email, err := client.Emails.Send(ctx, &rusender.SendEmailParams{
	From:    rusender.Address{Email: "no-reply@shop.com", Name: "Shop"},
	To:      []rusender.Address{rusender.Addr("client@example.com")},
	Subject: "Order placed",
	HTML:    "<p>Thank you for your order!</p>",
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(email.ID, email.Status) // 0199b2f4-…  queued
```

Create the client once and reuse it: it is safe for concurrent use by multiple goroutines.

Sending with a template:

```go
_, err = client.Emails.Send(ctx, &rusender.SendEmailParams{
	From:          rusender.Addr("no-reply@shop.com"),
	To:            []rusender.Address{{Email: "client@example.com", Name: "Anna"}},
	Subject:       "Order #1234",
	TemplateAlias: "order-created",
	Variables:     map[string]any{"orderId": 1234, "total": "$29.90"},
})
```

Attachments are bytes; the SDK encodes them to base64:

```go
invoice, err := os.ReadFile("invoice.pdf")
// ...
_, err = client.Emails.Send(ctx, &rusender.SendEmailParams{
	From:        rusender.Addr("no-reply@shop.com"),
	To:          []rusender.Address{rusender.Addr("client@example.com")},
	Subject:     "Invoice",
	Text:        "The invoice is attached",
	Attachments: []rusender.EmailAttachment{{FileName: "invoice.pdf", Content: invoice}},
})
```

A campaign with attachments: upload the files first (up to 20, 5 MB in total), then pass their `ID`s:

```go
price, _ := os.Open("price.pdf")
defer price.Close()

files, err := client.Files.Upload(ctx, []rusender.UploadFile{{FileName: "price.pdf", Content: price}})
// ...
campaign, err := client.Campaigns.Create(ctx, &rusender.CreateCampaignParams{
	Name:        "New price list",
	Subject:     "November prices",
	TemplateID:  12,
	SenderID:    3,
	ListIDs:     []int{7},
	Attachments: []int{files[0].ID},
})
```

Email status and recipient events:

```go
sent, err := client.Emails.Get(ctx, email.ID)
// ...
for _, recipient := range sent.Recipients {
	fmt.Println(recipient.Email, recipient.Status, recipient.Reason)
}
```

## Resources

| Resource | What it does |
|---|---|
| `Emails` | sending, lists of emails and deliveries, email status, recipient events |
| `Contacts` | search, update, unsubscribe and resubscribe, statistics, deletion, export, sending history |
| `Lists` | contact lists; `Lists.Contacts` — add, move, import, export; `Lists.Variables` — variables |
| `Campaigns` | campaigns: drafts, start, pause and resume, A/B tests, chunked sending, statistics, recipients |
| `Folders` | campaign folders |
| `Templates` | email templates |
| `Senders`, `Domains` | sender addresses and domains, DNS verification |
| `SendingKeys` | sending keys for transactional emails and their statistics |
| `Segments` | contact segments and conditions |
| `Webhooks` | event subscriptions: deliveries, opens, unsubscribes, campaign statuses; test event |
| `Files` | campaign attachment upload, file list, download |
| `Me` | the account and scopes of the current key |

Chains are not part of the SDK yet.

## Parameters

Every method takes a `context.Context` first. Parameters are a struct passed by pointer; `nil` means no
parameters.

Optional fields are plain values: an empty string, `0` and `false` are not sent. A field is a pointer only where
an empty value is meaningful and must be sendable: the contact name and phone or the template plain-text version
in `Update` (`Phone: rusender.String("")` clears the phone), and dates (`*time.Time`). Pointers have helpers
`rusender.String`, `rusender.Int`, `rusender.Bool`, `rusender.Time` and `rusender.Ptr`. Where the API accepts
`null` to clear a value, the field is a `rusender.Nullable`:

```go
_, err = client.Templates.Update(ctx, 12, &rusender.UpdateTemplateParams{
	Name:  "Order placed",
	Alias: rusender.Null[string](), // remove the alias; rusender.Value("order") sets it
})
```

In campaign and segment `Update`, an empty but non-`nil` slice (`ListIDs: []int{}`) is sent as `[]` and clears
the value; `nil` leaves it as is.

Enums are string types with constants (`rusender.CampaignDraft`, `rusender.ContactActive`). The types are
open: a new value from the API does not break decoding.

## Errors

Any non-2xx API response is a `*rusender.Error` with `StatusCode`, `Code`, `Message`, `Details`, `RequestID`
and `RetryAfter`. Include `RequestID` when you contact support. Check the error class with `errors.Is`:

```go
_, err := client.Emails.Send(ctx, params)
var apiErr *rusender.Error
switch {
case errors.Is(err, rusender.ErrValidation) && errors.As(err, &apiErr):
	log.Println(apiErr.Code, apiErr.Details) // [{Field: to, Reason: …}]
case errors.Is(err, rusender.ErrRateLimit) && errors.As(err, &apiErr):
	log.Printf("retry in %s", apiErr.RetryAfter)
case err != nil:
	return err
}
```

| Value for `errors.Is` | When |
|---|---|
| `ErrValidation` | 400, 422 — the request failed validation, fields are in `Details` |
| `ErrAuthentication` | 401 — the key is invalid or revoked |
| `ErrPaymentRequired` | 402 — not enough balance or plan limit |
| `ErrPermission` | 403 — the key lacks the required scope |
| `ErrNotFound` | 404 |
| `ErrConflict` | 409 — a conflict, or the `Idempotency-Key` was reused with a different body |
| `ErrRateLimit` | 429, `RetryAfter` — how long to wait before retrying |
| `*NetworkError` (`errors.As`) | the request did not get through or no response was received |
| `ErrTimeout` | no response within the timeout |
| `ErrUnexpectedResponse` | a response outside the API contract: an HTML page of a proxy with status 200, a list without `items` |

A cancelled context returns `ctx.Err()`: `context.Canceled` or `context.DeadlineExceeded`.

## The whole response

Methods return the data. To also get the `RequestID`, the rate limit and the `*http.Response` itself, pass
`rusender.WithResponseInto`:

```go
var info rusender.ResponseInfo
list, err := client.Lists.Get(ctx, 7, rusender.WithResponseInto(&info))
// ...
fmt.Println(info.RequestID, info.RateLimit) // &{Limit:120 Remaining:119 ResetAt:…}
// wait until the window resets: time.Until(info.RateLimit.ResetAt)
```

`info.Body` is the raw JSON of the response, including fields the SDK does not know yet.

## Dates

Moments in time in responses — `CreatedAt`, `UpdatedAt`, `ScheduledAt`, `SentAt` and others — are
`time.Time` (`*time.Time` when the value may be absent). Calendar days without time (`Date` in campaign
activity, `DateFrom`/`DateTo` in key statistics) stay strings. Dates in parameters are `time.Time` and are sent
to the API as ISO 8601:

```go
start := time.Date(2026, 12, 1, 9, 0, 0, 0, time.FixedZone("MSK", 3*60*60))
_, err = client.Campaigns.Schedule(ctx, campaignID, &rusender.ScheduleCampaignParams{ScheduledStartDate: &start})
```

## Retries and idempotency

The SDK retries on 502, 503, 504 and network errors — 2 times by default, with exponential backoff. POST,
PATCH and DELETE are retried safely: the SDK sends an `Idempotency-Key`, so the server does not perform the
request twice. Pass your own key — an order number, for example — as an option:

```go
_, err = client.Emails.Send(ctx, params, rusender.WithIdempotencyKey(fmt.Sprintf("order-%d", order.ID)))
```

When the rate limit is exceeded (429), the SDK neither waits nor retries; it returns `ErrRateLimit` right away.
How long to wait is in `apiErr.RetryAfter`: wait that long or queue the job. Pass the same idempotency key when
you retry.

## Pagination

Lists return one page:

```go
page, err := client.Contacts.List(ctx, &rusender.ListContactsParams{
	PageParams: rusender.PageParams{Page: 1, Limit: 100},
	Status:     []rusender.ContactStatus{rusender.ContactActive},
})
// ...
fmt.Println(len(page.Items), page.Pagination.TotalPages)
```

Emails and deliveries use a cursor:

```go
params := &rusender.ListEmailsParams{Limit: 100}
for {
	page, err := client.Emails.List(ctx, params)
	if err != nil {
		return err
	}
	// page.Items
	if page.NextCursor == nil {
		break
	}
	params.Cursor = *page.NextCursor
}
```

## Exports and files

`Contacts.Export`, `Lists.Contacts.Export` and `Files.Download` return an `*http.Response`: read or copy its
`Body` and close it. The timeout limits waiting for the response headers; reading the body is limited by the
context only.

```go
resp, err := client.Contacts.Export(ctx, &rusender.ExportContactsParams{Format: rusender.ExportCSV})
if err != nil {
	return err
}
defer resp.Body.Close()
_, err = io.Copy(file, resp.Body)
```

## Client settings

```go
client, err := rusender.NewClient(apiKey,
	rusender.WithBaseURL("https://api.rusender.ru/api"),
	rusender.WithTimeout(30*time.Second), // per attempt
	rusender.WithMaxRetries(2),
	rusender.WithHeader("X-Team", "billing"),
	rusender.WithHTTPClient(&http.Client{Transport: myTransport}),
)
```

Per-request options are the last arguments of any method: `WithIdempotencyKey`, `WithRequestID` (sent as
`X-Request-ID`), `WithRequestTimeout`, `WithRequestMaxRetries`, `WithRequestHeader`, `WithResponseInto`.

The full guide is in the [documentation](https://rusender.ru/docs/), the type reference is on
[pkg.go.dev](https://pkg.go.dev/github.com/rusender/rusender-go).

## License

MIT
