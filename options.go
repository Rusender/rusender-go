package rusender

import (
	"net/http"
	"strconv"
	"time"
)

const (
	// DefaultBaseURL is the RuSender API address.
	DefaultBaseURL = "https://api.rusender.ru/api"
	// DefaultTimeout is the timeout of a single attempt.
	DefaultTimeout = 30 * time.Second
	// DefaultMaxRetries is the number of retries on 502–504 and network errors.
	DefaultMaxRetries = 2
)

type clientConfig struct {
	baseURL    string
	timeout    time.Duration
	maxRetries int
	httpClient *http.Client
	headers    http.Header
}

// ClientOption configures a [Client].
type ClientOption func(*clientConfig)

// WithBaseURL sets the API address, https://api.rusender.ru/api by default.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *clientConfig) { c.baseURL = baseURL }
}

// WithHTTPClient sets the HTTP client: proxy, custom transport, logging, tests. Leave its Timeout at zero:
// the SDK applies its own timeout to every attempt.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *clientConfig) { c.httpClient = httpClient }
}

// WithTimeout sets the timeout of a single attempt, 30 seconds by default.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) { c.timeout = timeout }
}

// WithMaxRetries sets the number of retries on 502–504 and network errors, 2 by default. 429 is not retried.
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *clientConfig) { c.maxRetries = maxRetries }
}

// WithHeader adds a header sent with every request.
func WithHeader(key, value string) ClientOption {
	return func(c *clientConfig) { c.headers.Set(key, value) }
}

type requestConfig struct {
	idempotencyKey string
	requestID      string
	timeout        time.Duration
	maxRetries     *int
	headers        http.Header
	responseInto   *ResponseInfo
}

// RequestOption configures a single call.
type RequestOption func(*requestConfig)

// WithIdempotencyKey sets the Idempotency-Key of a POST, PATCH or DELETE request. Without it the SDK generates
// one, so the request can be retried safely on network failures. Format [a-zA-Z0-9._\-:]{1,64}, valid for 24 hours.
func WithIdempotencyKey(key string) RequestOption {
	return func(c *requestConfig) { c.idempotencyKey = key }
}

// WithRequestID sets an identifier for end-to-end tracing, sent as X-Request-ID.
func WithRequestID(requestID string) RequestOption {
	return func(c *requestConfig) { c.requestID = requestID }
}

// WithRequestTimeout overrides the timeout of a single attempt: until the response headers arrive for exports
// and downloads, until the whole response is read otherwise. Reading the body of an export or a download is
// limited by the context only.
func WithRequestTimeout(timeout time.Duration) RequestOption {
	return func(c *requestConfig) { c.timeout = timeout }
}

// WithRequestMaxRetries overrides the number of retries on 502–504 and network errors.
func WithRequestMaxRetries(maxRetries int) RequestOption {
	return func(c *requestConfig) { c.maxRetries = &maxRetries }
}

// WithRequestHeader adds a header to this request.
func WithRequestHeader(key, value string) RequestOption {
	return func(c *requestConfig) {
		if c.headers == nil {
			c.headers = http.Header{}
		}
		c.headers.Set(key, value)
	}
}

// WithResponseInto fills info with the request id, rate limit and raw response of a successful call:
//
//	var info rusender.ResponseInfo
//	contact, err := client.Contacts.Get(ctx, 1, rusender.WithResponseInto(&info))
//	fmt.Println(info.RequestID, info.RateLimit)
func WithResponseInto(info *ResponseInfo) RequestOption {
	return func(c *requestConfig) { c.responseInto = info }
}

// RateLimit is the rate limit of the API key from the X-RateLimit-* headers of a response.
type RateLimit struct {
	// Limit is the number of requests allowed in the current window.
	Limit int
	// Remaining is the number of requests left in the current window.
	Remaining int
	// ResetAt is when the window resets, from X-RateLimit-Reset (Unix seconds in the header).
	ResetAt time.Time
}

// ResponseInfo is what came with the data of a call, see [WithResponseInto].
type ResponseInfo struct {
	// RequestID is the server-side request identifier: include it when you contact support.
	RequestID string
	// RateLimit is nil when the response has no X-RateLimit-* headers.
	RateLimit *RateLimit
	// Response is the raw HTTP response; its body has already been read and closed.
	Response *http.Response
	// Body is the raw JSON of the response, including fields the SDK does not know yet.
	Body []byte
}

func readRateLimit(header http.Header) *RateLimit {
	if header.Get("X-RateLimit-Limit") == "" {
		return nil
	}
	limit, err1 := strconv.Atoi(header.Get("X-RateLimit-Limit"))
	remaining, err2 := strconv.Atoi(header.Get("X-RateLimit-Remaining"))
	reset, err3 := strconv.ParseInt(header.Get("X-RateLimit-Reset"), 10, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}
	return &RateLimit{Limit: limit, Remaining: remaining, ResetAt: time.Unix(reset, 0)}
}
