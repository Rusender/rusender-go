package rusender

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// recordSleeps makes retries instant and returns the delays the client asked for.
func recordSleeps(client *Client) *[]time.Duration {
	var delays []time.Duration
	client.Me.t.sleep = func(ctx context.Context, d time.Duration) error {
		delays = append(delays, d)
		return ctx.Err()
	}
	return &delays
}

var created = map[string]any{"id": 1, "name": "Clients"}

func TestPOSTKeepsIdempotencyKeyAcrossRetries(t *testing.T) {
	client, s := newStubClient(t, []step{fail(503, "UNAVAILABLE", "Down"), ok(created)})
	recordSleeps(client)
	if _, err := client.Lists.Create(context.Background(), &CreateListParams{Name: "Clients"}); err != nil {
		t.Fatal(err)
	}
	first, second := s.request(t, 0).Header.Get("Idempotency-Key"), s.request(t, 1).Header.Get("Idempotency-Key")
	if first == "" || first != second {
		t.Fatalf("keys %q and %q", first, second)
	}
	if len(first) != 36 || first[14] != '4' {
		t.Errorf("not a UUIDv4: %s", first)
	}
}

func TestCallerIdempotencyKey(t *testing.T) {
	client, s := newStubClient(t, []step{ok(created)})
	if _, err := client.Lists.Create(context.Background(), &CreateListParams{Name: "Clients"}, WithIdempotencyKey("order-42")); err != nil {
		t.Fatal(err)
	}
	if got := s.request(t, 0).Header.Get("Idempotency-Key"); got != "order-42" {
		t.Errorf("key %q", got)
	}
}

func TestRetriesInProgressConflict(t *testing.T) {
	client, s := newStubClient(t, []step{
		fail(409, "REQUEST_ALREADY_IN_PROGRESS", "In progress", "Retry-After", "2"),
		ok(created),
	})
	delays := recordSleeps(client)
	if _, err := client.Lists.Create(context.Background(), &CreateListParams{Name: "Clients"}); err != nil {
		t.Fatal(err)
	}
	if s.request(t, 0).Header.Get("Idempotency-Key") != s.request(t, 1).Header.Get("Idempotency-Key") {
		t.Error("the retry must reuse the key")
	}
	if len(*delays) != 1 || (*delays)[0] != 2*time.Second {
		t.Errorf("delays %v", *delays)
	}
}

func TestDoesNotRetryOtherErrors(t *testing.T) {
	for _, response := range []step{
		fail(409, "LIST_NAME_TAKEN", "Taken"),
		fail(500, "INTERNAL", "Oops"),
		fail(429, "RATE_LIMITED", "Slow", "Retry-After", "1"),
	} {
		client, s := newStubClient(t, []step{response})
		if _, err := client.Lists.Create(context.Background(), &CreateListParams{Name: "x"}); err == nil {
			t.Fatal("no error")
		}
		if s.count() != 1 {
			t.Errorf("%d requests, want 1", s.count())
		}
	}
}

func TestGivesUpAfterMaxRetries(t *testing.T) {
	client, s := newStubClient(t, []step{
		fail(502, "BAD_GATEWAY", "1"), fail(503, "UNAVAILABLE", "2"), fail(504, "TIMEOUT", "3"),
	})
	delays := recordSleeps(client)
	_, err := client.Lists.Get(context.Background(), 1)
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 504 {
		t.Fatalf("got %v, want the last error", err)
	}
	if s.count() != 3 || len(*delays) != 2 {
		t.Errorf("%d requests, %d delays", s.count(), len(*delays))
	}
	for i, d := range *delays {
		base := baseRetryDelay << i
		if d < base*3/4 || d > base*5/4 {
			t.Errorf("delay %d = %s, want about %s", i, d, base)
		}
	}
}

func TestRetryAfterOn503(t *testing.T) {
	client, _ := newStubClient(t, []step{fail(503, "UNAVAILABLE", "Down", "Retry-After", "3"), ok(created)})
	delays := recordSleeps(client)
	if _, err := client.Lists.Get(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if len(*delays) != 1 || (*delays)[0] != 3*time.Second {
		t.Errorf("delays %v", *delays)
	}
}

func TestLongRetryAfterIsNotWaited(t *testing.T) {
	client, s := newStubClient(t, []step{fail(503, "UNAVAILABLE", "Down", "Retry-After", "120")})
	recordSleeps(client)
	if _, err := client.Lists.Get(context.Background(), 1); err == nil {
		t.Fatal("no error")
	}
	if s.count() != 1 {
		t.Errorf("%d requests", s.count())
	}
}

func TestRetriesNetworkErrors(t *testing.T) {
	client, s := newStubClient(t, []step{networkError(errors.New("connection reset")), ok(created)})
	recordSleeps(client)
	if _, err := client.Lists.Get(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if s.count() != 2 {
		t.Errorf("%d requests", s.count())
	}
}

func TestFinalNetworkError(t *testing.T) {
	cause := errors.New("connection refused")
	client, _ := newStubClient(t, []step{networkError(cause)}, WithMaxRetries(0))
	_, err := client.Lists.Get(context.Background(), 1)
	var netErr *NetworkError
	if !errors.As(err, &netErr) || !errors.Is(err, cause) {
		t.Fatalf("got %v", err)
	}
}

func TestNonIdempotentOperationIsNotRetried(t *testing.T) {
	client, s := newStubClient(t, []step{fail(503, "UNAVAILABLE", "Down")})
	recordSleeps(client)
	if _, err := client.Webhooks.Test(context.Background(), 5, WithIdempotencyKey("ignored")); err == nil {
		t.Fatal("no error")
	}
	if s.count() != 1 {
		t.Errorf("%d requests", s.count())
	}
	if s.request(t, 0).Header.Get("Idempotency-Key") != "" {
		t.Error("Idempotency-Key must not be sent")
	}
}

func TestMaxRetriesPerRequest(t *testing.T) {
	client, s := newStubClient(t, []step{fail(503, "UNAVAILABLE", "Down")})
	recordSleeps(client)
	if _, err := client.Lists.Get(context.Background(), 1, WithRequestMaxRetries(0)); err == nil {
		t.Fatal("no error")
	}
	if s.count() != 1 {
		t.Errorf("%d requests", s.count())
	}
}

func hang(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}

func TestTimeout(t *testing.T) {
	client, _ := newStubClient(t, []step{hang}, WithTimeout(20*time.Millisecond), WithMaxRetries(0))
	_, err := client.Lists.Get(context.Background(), 1)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("got %v", err)
	}
}

func TestTimeoutIsRetried(t *testing.T) {
	client, s := newStubClient(t, []step{hang, ok(created)}, WithTimeout(20*time.Millisecond))
	recordSleeps(client)
	if _, err := client.Lists.Get(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if s.count() != 2 {
		t.Errorf("%d requests", s.count())
	}
}

func TestCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client, s := newStubClient(t, []step{func(req *http.Request) (*http.Response, error) {
		cancel()
		return hang(req)
	}})
	_, err := client.Lists.Get(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if s.count() != 1 {
		t.Errorf("a cancelled request must not be retried, %d requests", s.count())
	}
}

func TestAlreadyCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client, _ := newStubClient(t, []step{hang})
	if _, err := client.Lists.Get(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

// The timeout of a download covers the headers only; reading the body is limited by the context, and
// cancelling the context stops reading.
func TestDownloadBodyOutlivesTimeoutAndStopsOnCancel(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("part1,"))
		w.(http.Flusher).Flush()
		select {
		case <-release:
			_, _ = w.Write([]byte("part2"))
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	client, err := NewClient(testAPIKey, WithBaseURL(server.URL), WithTimeout(50*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Files.Download(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond) // past the timeout
	close(release)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || string(body) != "part1,part2" {
		t.Fatalf("body %q, err %v", body, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	release = make(chan struct{})
	resp, err = client.Files.Download(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	_, err = io.ReadAll(resp.Body)
	resp.Body.Close()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("reading after cancel: %v", err)
	}
}

func TestRealServerRoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/public/lists" || !strings.Contains(string(body), `"data":{"name":"Clients"}`) {
			http.Error(w, "unexpected request", http.StatusTeapot)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data":{"id":3,"name":"Clients","status":"enabled"},"meta":{"requestId":"r"}}`))
	}))
	defer server.Close()
	client, err := NewClient(testAPIKey, WithBaseURL(server.URL+"/api"))
	if err != nil {
		t.Fatal(err)
	}
	list, err := client.Lists.Create(context.Background(), &CreateListParams{Name: "Clients"})
	if err != nil {
		t.Fatal(err)
	}
	if list.ID != 3 || list.Status != StatusEnabled {
		t.Errorf("list %+v", list)
	}
}
