package boi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newTestClient(maxRetries int, rt http.RoundTripper) *Client {
	return &Client{
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		downloadClient: &http.Client{Transport: rt},
		maxRetries:     maxRetries,
	}
}

func makeResponse(status int, statusText string, body string, headers ...string) *http.Response {
	h := make(http.Header)
	for i := 0; i+1 < len(headers); i += 2 {
		h.Set(headers[i], headers[i+1])
	}
	return &http.Response{
		StatusCode: status,
		Status:     statusText,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestClientGetRetriesTransientAndRespectsRetryAfter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var attempts atomic.Int32
		client := newTestClient(4, roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			switch attempts.Add(1) {
			case 1:
				return makeResponse(http.StatusServiceUnavailable, "503 Service Unavailable", "busy", "Retry-After", "2"), nil
			case 2:
				return makeResponse(http.StatusServiceUnavailable, "503 Service Unavailable", "still busy"), nil
			default:
				return makeResponse(http.StatusOK, "200 OK", "payload"), nil
			}
		}))

		data, err := client.Get(t.Context(), "https://cdn.example/bundle")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if string(data) != "payload" {
			t.Fatalf("Get() body = %q, want payload", data)
		}
		if got := attempts.Load(); got != 3 {
			t.Fatalf("attempts = %d, want 3", got)
		}
	})
}

func TestClientGetStopsImmediatelyOnPermanentClientError(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var attempts atomic.Int32
		client := newTestClient(5, roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			attempts.Add(1)
			return makeResponse(http.StatusNotFound, "404 Not Found", "missing"), nil
		}))

		_, err := client.Get(t.Context(), "https://cdn.example/missing")
		if err == nil {
			t.Fatal("expected error for 404, got nil")
		}
		if got := attempts.Load(); got != 1 {
			t.Fatalf("attempts = %d, want 1 (404 should not be retried)", got)
		}
	})
}

func TestClientGetCancelsDuringBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var attempts atomic.Int32
		client := newTestClient(5, roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			attempts.Add(1)
			return makeResponse(http.StatusServiceUnavailable, "503 Service Unavailable", "busy", "Retry-After", "10"), nil
		}))

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		errCh := make(chan error, 1)
		go func() {
			_, err := client.Get(ctx, "https://cdn.example/slow")
			errCh <- err
		}()

		synctest.Wait()
		time.Sleep(50 * time.Millisecond)
		cancel()

		err := <-errCh
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Get() error = %v, want context.Canceled", err)
		}
		if got := attempts.Load(); got != 1 {
			t.Fatalf("attempts = %d, want 1", got)
		}
	})
}
