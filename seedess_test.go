package seedess

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateStream(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"id":"demo","publicToken":"pk_1","playback":{"signaling":"/s"}}`))
	}))
	defer server.Close()

	client := NewWithOptions("key_test", server.URL, nil)
	stream, err := client.CreateStream(context.Background(), CreateStreamInput{
		URL: "https://origin/master.m3u8", ID: "demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	if stream.ID != "demo" || stream.PublicToken != "pk_1" {
		t.Fatalf("unexpected stream: %+v", stream)
	}
	if gotPath != "/api/v1/streams" || gotAuth != "Bearer key_test" {
		t.Fatalf("path/auth: %q %q", gotPath, gotAuth)
	}
	if gotBody["id"] != "demo" {
		t.Fatalf("body: %+v", gotBody)
	}
}

func TestDisablePeersMapsToKill(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewWithOptions("key_test", server.URL, nil)
	if _, err := client.DisablePeers(context.Background(), "demo"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v1/streams/demo/kill" {
		t.Fatalf("path: %q", gotPath)
	}
}

func TestErrorEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"ok":false,"error":{"code":"QUOTA_EXCEEDED","message":"Limit reached"}}`))
	}))
	defer server.Close()

	client := NewWithOptions("key_test", server.URL, nil)
	_, err := client.CreateStream(context.Background(), CreateStreamInput{URL: "https://origin"})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Status != 429 || apiErr.Code != "QUOTA_EXCEEDED" {
		t.Fatalf("unexpected error: %+v", apiErr)
	}
}

func TestGetMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"id":"demo","metrics":{"cdnBytes":10,"p2pBytes":90,"offloadPercent":90}}`))
	}))
	defer server.Close()

	client := NewWithOptions("key_test", server.URL, nil)
	metrics, err := client.GetMetrics(context.Background(), "demo")
	if err != nil {
		t.Fatal(err)
	}
	if metrics.P2PBytes != 90 || metrics.OffloadPercent != 90 {
		t.Fatalf("unexpected metrics: %+v", metrics)
	}
}
