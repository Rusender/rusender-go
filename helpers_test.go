package rusender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
)

const testAPIKey = "rs_ck_v1_test_secret"

// recordedRequest is a request the stub received, with its body read.
type recordedRequest struct {
	*http.Request
	body []byte
}

// data returns the `data` of a JSON request body.
func (r recordedRequest) data(t *testing.T) map[string]any {
	t.Helper()
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(r.body, &envelope); err != nil {
		t.Fatalf("request body is not JSON: %v: %s", err, r.body)
	}
	return envelope.Data
}

type step func(*http.Request) (*http.Response, error)

// stub is an http.RoundTripper that answers with the steps in order and records the requests.
type stub struct {
	mu       sync.Mutex
	steps    []step
	requests []recordedRequest
}

func (s *stub) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}
	s.mu.Lock()
	s.requests = append(s.requests, recordedRequest{Request: req, body: body})
	if len(s.steps) == 0 {
		s.mu.Unlock()
		return nil, fmt.Errorf("unexpected request %s %s", req.Method, req.URL)
	}
	next := s.steps[0]
	s.steps = s.steps[1:]
	s.mu.Unlock()
	return next(req)
}

func (s *stub) request(t *testing.T, i int) recordedRequest {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if i >= len(s.requests) {
		t.Fatalf("request %d was not sent, %d sent", i, len(s.requests))
	}
	return s.requests[i]
}

func (s *stub) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.requests)
}

// newStubClient returns a client whose requests are answered by the steps.
func newStubClient(t *testing.T, steps []step, opts ...ClientOption) (*Client, *stub) {
	t.Helper()
	s := &stub{steps: steps}
	opts = append([]ClientOption{WithHTTPClient(&http.Client{Transport: s})}, opts...)
	client, err := NewClient(testAPIKey, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return client, s
}

func jsonResponse(status int, body any, header ...string) step {
	return func(req *http.Request) (*http.Response, error) {
		payload, _ := json.Marshal(body)
		return rawResponse(status, payload, append([]string{"Content-Type", "application/json"}, header...)...)(req)
	}
}

func rawResponse(status int, body []byte, header ...string) step {
	return func(req *http.Request) (*http.Response, error) {
		h := http.Header{}
		for i := 0; i+1 < len(header); i += 2 {
			h.Set(header[i], header[i+1])
		}
		return &http.Response{
			StatusCode: status,
			Header:     h,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    req,
		}, nil
	}
}

func ok(data any, meta ...map[string]any) step {
	return okStatus(http.StatusOK, data, meta...)
}

func okStatus(status int, data any, meta ...map[string]any) step {
	m := map[string]any{"requestId": "req_ok", "timestamp": "2026-11-11T00:00:00.000Z"}
	for _, extra := range meta {
		for k, v := range extra {
			m[k] = v
		}
	}
	return jsonResponse(status, map[string]any{"ok": true, "data": data, "meta": m})
}

func fail(status int, code, message string, header ...string) step {
	return jsonResponse(status, map[string]any{
		"ok":    false,
		"error": map[string]any{"code": code, "message": message},
		"meta":  map[string]any{"requestId": "req_err", "timestamp": "2026-11-11T00:00:00.000Z"},
	}, header...)
}

func networkError(err error) step {
	return func(*http.Request) (*http.Response, error) { return nil, err }
}

var testPagination = map[string]any{"pagination": map[string]any{"page": 1, "limit": 20, "totalItems": 0, "totalPages": 0}}
