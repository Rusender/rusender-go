// Package rusender is the official Go client for the RuSender public API.
//
// Create a client once and reuse it:
//
//	client, err := rusender.NewClient(os.Getenv("RUSENDER_API_KEY"))
//	if err != nil {
//		log.Fatal(err)
//	}
//	email, err := client.Emails.Send(ctx, &rusender.SendEmailParams{
//		From:    rusender.Address{Email: "no-reply@shop.ru"},
//		To:      []rusender.Address{{Email: "client@example.com"}},
//		Subject: "Hello",
//		Text:    "Hello!",
//	})
//
// Every method takes a [context.Context] first and returns an error that can be inspected with
// [errors.Is] and [errors.As]: API errors are [*Error], failed connections are [*NetworkError],
// a request without a response within the timeout matches [ErrTimeout].
//
// Optional request fields are plain values: an empty string, 0 and false are not sent. A field is a pointer
// only where an empty value is meaningful, such as the phone of a contact in an update; set it with [String],
// [Int], [Bool], [Time] or [Ptr]. A field that accepts null to clear a value is a [Nullable].
//
// Request id, rate limit and the raw response of a call are available through [WithResponseInto].
package rusender
