package rusender

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewClientRequiresAPIKey(t *testing.T) {
	t.Setenv("RUSENDER_API_KEY", "")
	if _, err := NewClient(""); !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("got %v, want ErrMissingAPIKey", err)
	}
}

func TestNewClientReadsAPIKeyFromEnv(t *testing.T) {
	t.Setenv("RUSENDER_API_KEY", "rs_ck_v1_from_env")
	s := &stub{steps: []step{ok(map[string]any{"accountId": 1})}}
	client, err := NewClient("", WithHTTPClient(&http.Client{Transport: s}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Me.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := s.request(t, 0).Header.Get("Authorization"); got != "Bearer rs_ck_v1_from_env" {
		t.Fatalf("Authorization %q", got)
	}
}

func TestDefaultHeadersAndBaseURL(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"accountId": 1})})
	if _, err := client.Me.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	req := s.request(t, 0)
	if req.URL.String() != "https://api.rusender.ru/api/v1/public/me" {
		t.Errorf("URL %s", req.URL)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer "+testAPIKey {
		t.Errorf("Authorization %q", got)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept %q", got)
	}
	if ua := req.Header.Get("User-Agent"); !regexp.MustCompile(`^rusender-go/\d+\.\d+\.\d+\S* \(go/[^)]+\)$`).MatchString(ua) {
		t.Errorf("User-Agent %q", ua)
	}
	if req.Header.Get("Idempotency-Key") != "" || req.Header.Get("Content-Type") != "" {
		t.Error("a GET must not send Idempotency-Key or Content-Type")
	}
}

func TestBaseURLHeadersAndRequestID(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"accountId": 1})},
		WithBaseURL("http://localhost:3000/api/"), WithHeader("X-Tenant", "shop"))
	_, err := client.Me.Get(context.Background(),
		WithRequestHeader("X-Debug", "1"), WithRequestID("trace-1"))
	if err != nil {
		t.Fatal(err)
	}
	req := s.request(t, 0)
	if req.URL.String() != "http://localhost:3000/api/v1/public/me" {
		t.Errorf("URL %s", req.URL)
	}
	for key, want := range map[string]string{"X-Tenant": "shop", "X-Debug": "1", "X-Request-ID": "trace-1"} {
		if got := req.Header.Get(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestQueryEncoding(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"items": []any{}}, testPagination)})
	_, err := client.Contacts.List(context.Background(), &ListContactsParams{
		PageParams: PageParams{Page: 2},
		Status:     []ContactStatus{ContactActive, ContactNew},
		Email:      "a+b@example.ru",
	})
	if err != nil {
		t.Fatal(err)
	}
	query := s.request(t, 0).URL.Query()
	if got := query["status"]; strings.Join(got, ",") != "active,new" {
		t.Errorf("status %v", got)
	}
	if query.Get("page") != "2" || query.Get("email") != "a+b@example.ru" {
		t.Errorf("query %v", query)
	}
	for _, absent := range []string{"limit", "listId", "sortBy", "sortDir"} {
		if query.Has(absent) {
			t.Errorf("%s must be left out", absent)
		}
	}
}

func TestQueryDatesAsISO(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"items": []any{}, "hasMore": false, "nextCursor": nil})})
	from := time.Date(2026, 11, 1, 3, 0, 0, 0, time.FixedZone("MSK", 3*3600))
	if _, err := client.Emails.List(context.Background(), &ListEmailsParams{DateFrom: &from}); err != nil {
		t.Fatal(err)
	}
	if got := s.request(t, 0).URL.Query().Get("dateFrom"); got != "2026-11-01T00:00:00.000Z" {
		t.Errorf("dateFrom %q", got)
	}
}

func TestPathParametersAreEscaped(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"id": "a/b"})})
	if _, err := client.Emails.Get(context.Background(), "a/b?c"); err != nil {
		t.Fatal(err)
	}
	if got := s.request(t, 0).URL.EscapedPath(); got != "/api/v1/public/emails/a%2Fb%3Fc" {
		t.Errorf("path %s", got)
	}
}

func TestVersionIsSemver(t *testing.T) {
	if !regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$`).MatchString(Version) {
		t.Fatalf("Version %q is not semver: the release tag is v + Version", Version)
	}
}

func TestAPIKeyDoesNotLeak(t *testing.T) {
	client, s := newStubClient(t, []step{ok(map[string]any{"accountId": 1})})
	for _, format := range []string{"%v", "%+v", "%#v", "%s"} {
		for _, value := range []any{client, client.Emails, client.Lists.Contacts, client.Emails.t} {
			if out := fmt.Sprintf(format, value); strings.Contains(out, testAPIKey) {
				t.Errorf("%s of %T leaks the key: %s", format, value, out)
			}
		}
	}
	if _, err := client.Me.Get(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := s.request(t, 0).Header.Get("Authorization"); got != "Bearer "+testAPIKey {
		t.Errorf("Authorization %q", got)
	}
}

func TestInvalidSettings(t *testing.T) {
	for name, opt := range map[string]ClientOption{
		"zero timeout":     WithTimeout(0),
		"negative timeout": WithTimeout(-time.Second),
		"negative retries": WithMaxRetries(-1),
	} {
		if _, err := NewClient(testAPIKey, opt); err == nil {
			t.Errorf("%s: no error", name)
		}
	}

	client, s := newStubClient(t, nil)
	if _, err := client.Me.Get(context.Background(), WithRequestMaxRetries(-1)); err == nil {
		t.Error("negative per-request retries: no error")
	}
	if s.count() != 0 {
		t.Error("an invalid request must not be sent")
	}
}

func TestAPIErrors(t *testing.T) {
	cases := []struct {
		status int
		class  error
	}{
		{400, ErrValidation}, {422, ErrValidation}, {401, ErrAuthentication}, {402, ErrPaymentRequired},
		{403, ErrPermission}, {404, ErrNotFound}, {409, ErrConflict}, {429, ErrRateLimit},
	}
	for _, c := range cases {
		client, _ := newStubClient(t, []step{fail(c.status, "SOME_CODE", "Something")})
		_, err := client.Lists.Get(context.Background(), 1)
		if !errors.Is(err, c.class) {
			t.Errorf("%d: %v is not %v", c.status, err, c.class)
		}
		for _, other := range cases {
			if other.class != c.class && errors.Is(err, other.class) {
				t.Errorf("%d also matches %v", c.status, other.class)
			}
		}
	}
}

func TestAPIErrorFields(t *testing.T) {
	client, _ := newStubClient(t, []step{jsonResponse(422, map[string]any{
		"ok": false,
		"error": map[string]any{
			"code":    "VALIDATION_ERROR",
			"message": "Invalid email",
			"details": []any{map[string]any{"field": "email", "reason": "invalid", "value": "x"}},
		},
		"meta": map[string]any{"requestId": "req_1"},
	})})
	_, err := client.Lists.Contacts.Create(context.Background(), 1, &CreateListContactParams{Email: "x"})
	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("%v is not *Error", err)
	}
	if apiErr.StatusCode != 422 || apiErr.Code != "VALIDATION_ERROR" || apiErr.Message != "Invalid email" ||
		apiErr.RequestID != "req_1" || len(apiErr.Details) != 1 || apiErr.Details[0].Field != "email" {
		t.Errorf("unexpected error %+v", apiErr)
	}
	if !strings.Contains(err.Error(), "Invalid email") || !strings.Contains(err.Error(), "req_1") {
		t.Errorf("message %q", err.Error())
	}
}

func TestRateLimitRetryAfter(t *testing.T) {
	client, s := newStubClient(t, []step{fail(429, "RATE_LIMITED", "Slow down", "Retry-After", "7")})
	_, err := client.Lists.Get(context.Background(), 1)
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.RetryAfter != 7*time.Second {
		t.Fatalf("got %v", err)
	}
	if s.count() != 1 {
		t.Errorf("429 must not be retried, %d requests", s.count())
	}
}

func TestNonJSONErrorBody(t *testing.T) {
	client, _ := newStubClient(t, []step{rawResponse(500, []byte("<html>Internal error</html>"))})
	_, err := client.Lists.Get(context.Background(), 1)
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Code != "HTTP_500" || !strings.Contains(apiErr.Message, "<html>Internal error") {
		t.Fatalf("got %#v", err)
	}
}

func TestMissingRouteKeepsServerMessage(t *testing.T) {
	client, _ := newStubClient(t, []step{jsonResponse(404, map[string]any{
		"message": "Cannot GET /api/v1/public/nope", "error": "Not Found", "statusCode": 404,
	})})
	_, err := client.Lists.Get(context.Background(), 1)
	var apiErr *Error
	if !errors.As(err, &apiErr) || !strings.Contains(apiErr.Message, "Cannot GET /api/v1/public/nope") {
		t.Fatalf("got %v", err)
	}
}

func TestResponseOutsideContract(t *testing.T) {
	cases := map[string]step{
		"HTML page":             rawResponse(200, []byte("<html>Login to Wi-Fi</html>"), "Content-Type", "text/html"),
		"JSON without envelope": jsonResponse(200, map[string]any{"id": 1}),
		"ok: false with 200":    jsonResponse(200, map[string]any{"ok": false}),
		"empty body":            rawResponse(200, nil),
	}
	for name, response := range cases {
		client, _ := newStubClient(t, []step{response})
		if _, err := client.Lists.Get(context.Background(), 1); !errors.Is(err, ErrUnexpectedResponse) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

func TestListWithoutItemsIsAnError(t *testing.T) {
	client, _ := newStubClient(t, []step{ok(map[string]any{}, testPagination)})
	if _, err := client.Lists.List(context.Background(), nil); !errors.Is(err, ErrUnexpectedResponse) {
		t.Fatalf("got %v", err)
	}
	client, _ = newStubClient(t, []step{ok(map[string]any{"items": []any{}})})
	if _, err := client.Lists.List(context.Background(), nil); !errors.Is(err, ErrUnexpectedResponse) {
		t.Fatalf("page without pagination: got %v", err)
	}
}

func TestWrongTypeIsUnexpectedResponse(t *testing.T) {
	client, _ := newStubClient(t, []step{ok(map[string]any{"id": "not a number"})})
	if _, err := client.Lists.Get(context.Background(), 1); !errors.Is(err, ErrUnexpectedResponse) {
		t.Fatalf("got %v", err)
	}
}

func TestWithResponseInto(t *testing.T) {
	client, s := newStubClient(t, []step{jsonResponse(200, map[string]any{
		"ok": true, "data": map[string]any{"id": 7, "name": "Clients", "futureField": 1}, "meta": map[string]any{"requestId": "req_7"},
	}, "X-RateLimit-Limit", "100", "X-RateLimit-Remaining", "99", "X-RateLimit-Reset", "1795000000")})
	var info ResponseInfo
	list, err := client.Lists.Get(context.Background(), 7, WithResponseInto(&info))
	if err != nil {
		t.Fatal(err)
	}
	if list.Name != "Clients" || info.RequestID != "req_7" || info.Response == nil || info.Response.StatusCode != 200 {
		t.Errorf("list %+v, info %+v", list, info)
	}
	if info.RateLimit == nil || info.RateLimit.Limit != 100 || info.RateLimit.Remaining != 99 ||
		!info.RateLimit.ResetAt.Equal(time.Unix(1795000000, 0)) {
		t.Errorf("rate limit %+v", info.RateLimit)
	}
	if !strings.Contains(string(info.Body), "futureField") {
		t.Error("Body must keep fields the SDK does not know")
	}
	if s.count() != 1 {
		t.Errorf("%d requests", s.count())
	}
}

func TestWithResponseIntoWithoutRateLimit(t *testing.T) {
	client, _ := newStubClient(t, []step{ok(map[string]any{"id": 7})})
	var info ResponseInfo
	if _, err := client.Lists.Get(context.Background(), 7, WithResponseInto(&info)); err != nil {
		t.Fatal(err)
	}
	if info.RateLimit != nil {
		t.Errorf("rate limit %+v", info.RateLimit)
	}
}

func TestResponseDatesAreTime(t *testing.T) {
	client, _ := newStubClient(t, []step{ok(map[string]any{
		"id": 7, "createdAt": "2026-11-12T12:00:00.000Z", "updatedAt": "2026-11-12T13:00:00.000Z",
	})})
	list, err := client.Lists.Get(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if !list.UpdatedAt.Equal(time.Date(2026, 11, 12, 13, 0, 0, 0, time.UTC)) {
		t.Errorf("updatedAt %s", list.UpdatedAt)
	}
}

func TestDownloadReturnsBody(t *testing.T) {
	client, _ := newStubClient(t, []step{rawResponse(200, []byte("email\na@example.ru\n"), "Content-Type", "text/csv")})
	resp, err := client.Contacts.Export(context.Background(), &ExportContactsParams{Format: ExportCSV})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "email\na@example.ru\n" {
		t.Errorf("body %q", body)
	}
}

func TestFailedDownloadIsAnError(t *testing.T) {
	client, _ := newStubClient(t, []step{fail(404, "FILE_NOT_FOUND", "File not found")})
	if _, err := client.Files.Download(context.Background(), 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestAPIKeyIsTrimmed(t *testing.T) {
	for _, key := range []string{testAPIKey + "\n", "  " + testAPIKey + " ", `"` + testAPIKey + `"`, "'" + testAPIKey + "'"} {
		s := &stub{steps: []step{ok(map[string]any{"accountId": 1})}}
		client, err := NewClient(key, WithHTTPClient(&http.Client{Transport: s}))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Me.Get(context.Background()); err != nil {
			t.Fatal(err)
		}
		if got := s.request(t, 0).Header.Get("Authorization"); got != "Bearer "+testAPIKey {
			t.Errorf("key %q: Authorization %q", key, got)
		}
	}

	t.Setenv("RUSENDER_API_KEY", testAPIKey+"\n")
	if _, err := NewClient("  "); err != nil {
		t.Errorf("a blank argument falls back to the environment: %v", err)
	}
	t.Setenv("RUSENDER_API_KEY", " \n")
	if _, err := NewClient(""); !errors.Is(err, ErrMissingAPIKey) {
		t.Errorf("a blank key: got %v", err)
	}
}
