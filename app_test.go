package main

import (
	"context"
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
