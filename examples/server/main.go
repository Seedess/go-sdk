// Full working Seedess example for Go.
//
//	SEEDESS_API_KEY=key_... go run ./examples/server
//	open http://localhost:8787
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	seedess "github.com/Seedess/go-sdk"
)

//go:embed public/index.html
var indexHTML []byte

const origin = "https://seedess.com"

var (
	mu     sync.Mutex
	cached map[string]any
)

func handoff(client *seedess.Client) (map[string]any, error) {
	mu.Lock()
	defer mu.Unlock()
	if cached != nil {
		return cached, nil
	}

	streamURL := os.Getenv("SEEDESS_STREAM_URL")
	if streamURL == "" {
		streamURL = "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8"
	}
	streamID := os.Getenv("SEEDESS_STREAM_ID")
	if streamID == "" {
		streamID = "sdk-example"
	}

	stream, err := client.CreateStream(context.Background(), seedess.CreateStreamInput{
		URL: streamURL,
		ID:  streamID,
	})
	if err != nil {
		return nil, err
	}

	token := stream.PublicToken
	cached = map[string]any{
		"ok":     true,
		"stream": stream.ID,
		"token":  token,
		"playback": map[string]string{
			"master":    origin + stream.Playback.Master,
			"signaling": origin + stream.Playback.Signaling,
			"config":    origin + "/signaling/config?token=" + url.QueryEscape(token),
		},
	}
	return cached, nil
}

func main() {
	apiKey := os.Getenv("SEEDESS_API_KEY")
	if apiKey == "" {
		log.Fatal("Set SEEDESS_API_KEY (a server key, key_...) before running this example.")
	}
	baseURL := os.Getenv("SEEDESS_API_BASE")
	if baseURL == "" {
		baseURL = origin
	}
	client := seedess.NewWithOptions(apiKey, baseURL, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/seedess", func(w http.ResponseWriter, r *http.Request) {
		payload, err := handoff(client)
		w.Header().Set("content-type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(payload)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8787"
	}
	log.Printf("Seedess example on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, mux))
}
