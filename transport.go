package rusender

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	maxRetryAfter   = 60 * time.Second
	baseRetryDelay  = 500 * time.Millisecond
	maxRetryDelay   = 8 * time.Second
	queryTimeLayout = "2006-01-02T15:04:05.000Z07:00"
)

// request describes one API operation.
type request struct {
	method string
	path   string
	// query is a struct whose json tags name the query parameters.
	query any
	// body is the payload, sent to the API as {"data": body}; nil means no body.
	body any
	// form is a multipart/form-data body instead of JSON, for file uploads.
	form *multipartBody
	// notIdempotent marks a mutating operation without Idempotency-Key support: the key is not sent and the
	// request is not retried, since a retry could perform the action twice.
	notIdempotent bool
}

type multipartBody struct {
	contentType string
	data        []byte
}

type responseMeta struct {
	RequestID  string      `json:"requestId"`
	Pagination *Pagination `json:"pagination"`
}

type envelope struct {
	OK    *bool           `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Meta  *responseMeta   `json:"meta"`
	Error *struct {
		Code    json.RawMessage `json:"code"`
		Message json.RawMessage `json:"message"`
		Details json.RawMessage `json:"details"`
	} `json:"error"`
}

// transport sends requests: envelope, retries, idempotency, errors.
type transport struct {
	// a closure, not a string: fmt prints unexported fields of nested structs as they are, a func only as
	// its address, so the key cannot reach logs through fmt.Printf("%+v", client.Emails)
	apiKey    func() string
	cfg       clientConfig
	userAgent string
	// sleep waits between attempts; tests replace it to record the delays without waiting
	sleep func(ctx context.Context, d time.Duration) error
}

func newTransport(apiKey string, cfg clientConfig) *transport {
	return &transport{
		apiKey:    func() string { return apiKey },
		cfg:       cfg,
		userAgent: fmt.Sprintf("rusender-go/%s (go/%s)", Version, strings.TrimPrefix(runtime.Version(), "go")),
		sleep:     sleep,
	}
}

func (t *transport) String() string   { return "rusender.transport{apiKey: [hidden]}" }
func (t *transport) GoString() string { return t.String() }

// result is the parsed envelope of a successful JSON response.
type result struct {
	data json.RawMessage
	meta responseMeta
}

// do sends a request whose response is the JSON envelope.
func (t *transport) do(ctx context.Context, req request, opts []RequestOption) (*result, error) {
	rc := t.requestConfig(opts)
	resp, body, err := t.send(ctx, req, rc, true)
	if err != nil {
		return nil, err
	}
	// A 2xx that is not our envelope (an HTML page of a proxy, a captive portal) is not data:
	// returning it would hand the caller zero values under the API's types
	var env envelope
	if json.Unmarshal(body, &env) != nil || env.OK == nil || !*env.OK {
		return nil, fmt.Errorf("%w (HTTP %d): %s", ErrUnexpectedResponse, resp.StatusCode, describeBody(body))
	}
	res := &result{data: env.Data}
	if env.Meta != nil {
		res.meta = *env.Meta
	}
	if rc.responseInto != nil {
		requestID := res.meta.RequestID
		if requestID == "" {
			requestID = resp.Header.Get("X-Request-ID")
		}
		*rc.responseInto = ResponseInfo{
			RequestID: requestID,
			RateLimit: readRateLimit(resp.Header),
			Response:  resp,
			Body:      body,
		}
	}
	return res, nil
}

// raw sends a request whose response is a file or a stream (exports, downloads). The caller closes the body.
func (t *transport) raw(ctx context.Context, req request, opts []RequestOption) (*http.Response, error) {
	resp, _, err := t.send(ctx, req, t.requestConfig(opts), false)
	return resp, err
}

func (t *transport) requestConfig(opts []RequestOption) *requestConfig {
	rc := &requestConfig{}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}

func (t *transport) send(ctx context.Context, req request, rc *requestConfig, readBody bool) (*http.Response, []byte, error) {
	mutating := req.method != http.MethodGet
	idempotencyKey := ""
	if mutating && !req.notIdempotent {
		idempotencyKey = rc.idempotencyKey
		if idempotencyKey == "" {
			idempotencyKey = newUUID()
		}
	}
	safeToRetry := !mutating || idempotencyKey != ""

	timeout := t.cfg.timeout
	if rc.timeout != 0 {
		timeout = rc.timeout
	}
	maxRetries := t.cfg.maxRetries
	if rc.maxRetries != nil {
		maxRetries = *rc.maxRetries
	}
	if err := validateSettings(timeout, maxRetries); err != nil {
		return nil, nil, err
	}

	target, err := t.buildURL(req)
	if err != nil {
		return nil, nil, err
	}
	payload, contentType, err := encodeBody(req)
	if err != nil {
		return nil, nil, err
	}
	header := t.buildHeader(rc, contentType, idempotencyKey)

	for attempt := 0; ; attempt++ {
		canRetry := attempt < maxRetries
		attemptCtx, cancel := context.WithCancel(ctx)
		var timedOut atomic.Bool
		timer := time.AfterFunc(timeout, func() {
			timedOut.Store(true)
			cancel()
		})

		var bodyReader io.Reader
		if payload != nil {
			bodyReader = bytes.NewReader(payload)
		}
		httpReq, err := http.NewRequestWithContext(attemptCtx, req.method, target, bodyReader)
		if err != nil {
			timer.Stop()
			cancel()
			return nil, nil, err
		}
		httpReq.Header = header.Clone()

		resp, err := t.cfg.httpClient.Do(httpReq)
		var body []byte
		ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300
		if err == nil && (readBody || !ok) {
			body, err = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
		if err != nil {
			timer.Stop()
			cancel()
			if ctx.Err() != nil {
				return nil, nil, ctx.Err()
			}
			var failure error
			if timedOut.Load() {
				failure = fmt.Errorf("%w after %s", ErrTimeout, timeout)
			} else {
				failure = &NetworkError{Err: err}
			}
			if canRetry && safeToRetry {
				if err := t.sleep(ctx, backoff(attempt)); err != nil {
					return nil, nil, err
				}
				continue
			}
			return nil, nil, failure
		}
		timer.Stop()

		if ok {
			if readBody {
				cancel()
			} else {
				// The body of an export or a download is read by the caller after we return: the attempt's
				// context must live until then, so that cancelling ctx also stops reading
				resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
			}
			return resp, body, nil
		}
		cancel()

		apiErr := toAPIError(resp, body)
		retryable := isRetryable(resp.StatusCode, apiErr.Code, safeToRetry)
		// The server says how long to wait for 409 in progress and for 502–504; a longer wait is not ours to make
		if canRetry && retryable && apiErr.RetryAfter <= maxRetryAfter {
			delay := apiErr.RetryAfter
			if resp.Header.Get("Retry-After") == "" {
				delay = backoff(attempt)
			}
			if err := t.sleep(ctx, delay); err != nil {
				return nil, nil, err
			}
			continue
		}
		return nil, nil, apiErr
	}
}

type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnClose) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

func (t *transport) buildURL(req request) (string, error) {
	target := strings.TrimRight(t.cfg.baseURL, "/") + req.path
	if req.query == nil {
		return target, nil
	}
	values := url.Values{}
	if err := encodeQuery(values, reflect.ValueOf(req.query)); err != nil {
		return "", err
	}
	if len(values) == 0 {
		return target, nil
	}
	return target + "?" + values.Encode(), nil
}

func (t *transport) buildHeader(rc *requestConfig, contentType, idempotencyKey string) http.Header {
	header := t.cfg.headers.Clone()
	if header == nil {
		header = http.Header{}
	}
	header.Set("Authorization", "Bearer "+t.apiKey())
	header.Set("Accept", "application/json")
	header.Set("User-Agent", t.userAgent)
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	if idempotencyKey != "" {
		header.Set("Idempotency-Key", idempotencyKey)
	}
	if rc.requestID != "" {
		header.Set("X-Request-ID", rc.requestID)
	}
	for key, values := range rc.headers {
		header[key] = values
	}
	return header
}

func encodeBody(req request) ([]byte, string, error) {
	if req.form != nil {
		return req.form.data, req.form.contentType, nil
	}
	if req.body == nil {
		return nil, "", nil
	}
	body := req.body
	// a nil parameters pointer means "no parameters": the API gets an empty object, not null
	if v := reflect.ValueOf(body); v.Kind() == reflect.Pointer && v.IsNil() {
		body = struct{}{}
	}
	payload, err := json.Marshal(struct {
		Data any `json:"data"`
	}{body})
	if err != nil {
		return nil, "", fmt.Errorf("rusender: encode request body: %w", err)
	}
	return payload, "application/json", nil
}

var timeType = reflect.TypeOf(time.Time{})

// encodeQuery writes the fields of a parameters struct as query parameters, named by their json tags.
// Nil pointers, nil slices and zero values of omitempty fields are left out; slices repeat the parameter.
func encodeQuery(values url.Values, v reflect.Value) error {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("rusender: query parameters must be a struct, got %s", v.Type())
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i)
		if !field.IsExported() {
			continue
		}
		if field.Anonymous {
			if err := encodeQuery(values, value); err != nil {
				return err
			}
			continue
		}
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" || name == "" {
			continue
		}
		if value.Kind() == reflect.Pointer {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
		} else if strings.Contains(options, "omitempty") && value.IsZero() {
			continue
		}
		if value.Kind() == reflect.Slice {
			for j := 0; j < value.Len(); j++ {
				values.Add(name, queryScalar(value.Index(j)))
			}
			continue
		}
		values.Add(name, queryScalar(value))
	}
	return nil
}

func queryScalar(v reflect.Value) string {
	if v.Type() == timeType {
		return v.Interface().(time.Time).UTC().Format(queryTimeLayout)
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	}
	return fmt.Sprint(v.Interface())
}

func describeBody(body []byte) string {
	if len(bytes.TrimSpace(body)) == 0 {
		return "empty body"
	}
	if json.Valid(body) {
		return "JSON without the {ok, data} envelope"
	}
	return truncate(string(body), 200)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// plainMessage is the message of a JSON error that is not our envelope, e.g. NestJS {"message": "Cannot POST /…"}.
func plainMessage(body []byte) string {
	var plain struct {
		Message json.RawMessage `json:"message"`
	}
	if json.Unmarshal(body, &plain) != nil || plain.Message == nil {
		return ""
	}
	if text := jsonString(plain.Message); text != "" {
		return text
	}
	var parts []any
	if json.Unmarshal(plain.Message, &parts) == nil {
		texts := make([]string, len(parts))
		for i, part := range parts {
			texts[i] = fmt.Sprint(part)
		}
		return strings.Join(texts, "; ")
	}
	return ""
}

func jsonString(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return ""
	}
	return text
}

func toAPIError(resp *http.Response, body []byte) *Error {
	apiErr := &Error{StatusCode: resp.StatusCode, Header: resp.Header}
	var env envelope
	isEnvelope := json.Unmarshal(body, &env) == nil && env.OK != nil
	if isEnvelope && env.Error != nil {
		apiErr.Code = jsonString(env.Error.Code)
		apiErr.Message = jsonString(env.Error.Message)
		if env.Error.Details != nil {
			_ = json.Unmarshal(env.Error.Details, &apiErr.Details)
		}
	}
	if apiErr.Code == "" {
		apiErr.Code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
	}
	if apiErr.Message == "" {
		apiErr.Message = fmt.Sprintf("RuSender API responded with HTTP %d", resp.StatusCode)
		if !isEnvelope {
			if fallback := plainMessage(body); fallback != "" {
				apiErr.Message += ": " + fallback
			} else if len(bytes.TrimSpace(body)) > 0 && !json.Valid(body) {
				apiErr.Message += ": " + truncate(string(body), 200)
			}
		}
	}
	if isEnvelope && env.Meta != nil {
		apiErr.RequestID = env.Meta.RequestID
	}
	if apiErr.RequestID == "" {
		apiErr.RequestID = resp.Header.Get("X-Request-ID")
	}
	apiErr.RetryAfter = retryAfter(resp.Header)
	return apiErr
}

// isRetryable: 502–504 and 409 REQUEST_ALREADY_IN_PROGRESS. 429 is not retried: whether to wait for
// Retry-After is up to the caller (Error.RetryAfter).
func isRetryable(status int, code string, safeToRetry bool) bool {
	if !safeToRetry {
		return false
	}
	if status == http.StatusConflict {
		return code == "REQUEST_ALREADY_IN_PROGRESS"
	}
	return status == http.StatusBadGateway || status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout
}

func retryAfter(header http.Header) time.Duration {
	value := header.Get("Retry-After")
	if value == "" {
		return 0
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		return max(0, time.Duration(seconds*float64(time.Second)))
	}
	if date, err := http.ParseTime(value); err == nil {
		return max(0, time.Until(date))
	}
	return 0
}

func backoff(attempt int) time.Duration {
	delay := maxRetryDelay
	if attempt < 5 {
		delay = min(maxRetryDelay, baseRetryDelay<<attempt)
	}
	return time.Duration(float64(delay) * (0.75 + mathrand.Float64()*0.5))
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// validateSettings rejects settings that would break silently.
func validateSettings(timeout time.Duration, maxRetries int) error {
	if timeout <= 0 {
		return fmt.Errorf("rusender: timeout must be positive, got %s", timeout)
	}
	if maxRetries < 0 {
		return fmt.Errorf("rusender: maxRetries must be non-negative, got %d", maxRetries)
	}
	return nil
}

// pathf builds a path from a format with %s verbs, escaping every parameter.
func pathf(format string, params ...any) string {
	escaped := make([]any, len(params))
	for i, param := range params {
		escaped[i] = url.PathEscape(fmt.Sprint(param))
	}
	return fmt.Sprintf(format, escaped...)
}

// decode unmarshals the data of a response; a mismatch with the expected type is a broken contract.
func decode[T any](req request, data json.RawMessage) (T, error) {
	var out T
	if len(data) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, fmt.Errorf("%w: %s %s: %v", ErrUnexpectedResponse, req.method, req.path, err)
	}
	return out, nil
}

func call[T any](ctx context.Context, t *transport, req request, opts []RequestOption) (*T, error) {
	res, err := t.do(ctx, req, opts)
	if err != nil {
		return nil, err
	}
	out, err := decode[T](req, res.data)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// itemsOf requires an items array: anything else is a broken contract, not an empty list.
func itemsOf[T any](req request, data json.RawMessage) ([]T, error) {
	holder, err := decode[struct {
		Items *[]T `json:"items"`
	}](req, data)
	if err != nil {
		return nil, err
	}
	if holder.Items == nil {
		return nil, fmt.Errorf("%w: %s %s returned no items array", ErrUnexpectedResponse, req.method, req.path)
	}
	return *holder.Items, nil
}

// listPage is a page/limit list: items from data, pagination from meta.pagination.
func listPage[T any](ctx context.Context, t *transport, req request, opts []RequestOption) (*Page[T], error) {
	res, err := t.do(ctx, req, opts)
	if err != nil {
		return nil, err
	}
	items, err := itemsOf[T](req, res.data)
	if err != nil {
		return nil, err
	}
	if res.meta.Pagination == nil {
		return nil, fmt.Errorf("%w: %s %s returned no pagination", ErrUnexpectedResponse, req.method, req.path)
	}
	return &Page[T]{Items: items, Pagination: *res.meta.Pagination}, nil
}

func listCursorPage[T any](ctx context.Context, t *transport, req request, opts []RequestOption) (*CursorPage[T], error) {
	res, err := t.do(ctx, req, opts)
	if err != nil {
		return nil, err
	}
	items, err := itemsOf[T](req, res.data)
	if err != nil {
		return nil, err
	}
	page, err := decode[CursorPage[T]](req, res.data)
	if err != nil {
		return nil, err
	}
	page.Items = items
	return &page, nil
}

// listItems is a list without pagination: the API returns all items at once.
func listItems[T any](ctx context.Context, t *transport, req request, opts []RequestOption) ([]T, error) {
	res, err := t.do(ctx, req, opts)
	if err != nil {
		return nil, err
	}
	return itemsOf[T](req, res.data)
}
