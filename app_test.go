package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

// readyHandler reads a package-level global, so these subtests mutate it
// directly to exercise both branches; they intentionally don't use
// t.Parallel() to avoid racing on that shared state.
func Test_readyHandler(t *testing.T) {
	tests := []struct {
		name           string
		isReady        bool
		wantStatusCode int
	}{
		{
			name:           "ready",
			isReady:        true,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "not ready",
			isReady:        false,
			wantStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isReady = test.isReady

			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ready", http.NoBody)

			readyHandler(rec, req)

			if rec.Code != test.wantStatusCode {
				t.Errorf("readyHandler() status code = %d want %d", rec.Code, test.wantStatusCode)
			}
		})
	}
}

func Test_withHSTS(t *testing.T) {
	tests := []struct {
		name              string
		enabled           bool
		tls               bool
		forwardedProto    string
		wantHeaderPresent bool
	}{
		{
			name:              "disabled",
			enabled:           false,
			tls:               true,
			wantHeaderPresent: false,
		},
		{
			name:              "enabled, direct TLS",
			enabled:           true,
			tls:               true,
			wantHeaderPresent: true,
		},
		{
			name:              "enabled, X-Forwarded-Proto https",
			enabled:           true,
			forwardedProto:    "https",
			wantHeaderPresent: true,
		},
		{
			name:              "enabled, plain HTTP",
			enabled:           true,
			wantHeaderPresent: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody)
			if test.tls {
				req.TLS = &tls.ConnectionState{}
			}
			if test.forwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", test.forwardedProto)
			}

			withHSTS(next, test.enabled).ServeHTTP(rec, req)

			_, got := rec.Result().Header["Strict-Transport-Security"]
			if got != test.wantHeaderPresent {
				t.Errorf("Strict-Transport-Security present = %v want %v", got, test.wantHeaderPresent)
			}
		})
	}
}
