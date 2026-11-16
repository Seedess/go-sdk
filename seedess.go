// Package seedess is the official Seedess SDK for Go.
//
// Server-side only: it uses a secret API key (key_…) to create and control
// streams. Playback uses the separate @seedess/*-p2p packages and a per-stream
// public token (pk_…).
package seedess

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultBaseURL = "https://seedess.com"

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

type APIError struct {
	Status  int
	Code    string
	Message string
	Body    map[string]any
}

func (e *APIError) Error() string {
	return fmt.Sprintf("seedess: %s (%d): %s", e.Code, e.Status, e.Message)
}

type Playback struct {
	Format    string `json:"format,omitempty"`
	Master    string `json:"master,omitempty"`
	Manifest  string `json:"manifest,omitempty"`
	Playlist  string `json:"playlist,omitempty"`
	Signaling string `json:"signaling,omitempty"`
	Config    string `json:"config,omitempty"`
}

type Stream struct {
	ID             string           `json:"id"`
	Title          string           `json:"title,omitempty"`
	Live           bool             `json:"live,omitempty"`
	Status         string           `json:"status,omitempty"`
	Format         string           `json:"format,omitempty"`
	KillSwitch     bool             `json:"killSwitch,omitempty"`
	TargetDuration float64          `json:"targetDuration,omitempty"`
	MediaSequence  int              `json:"mediaSequence,omitempty"`
	Variants       []map[string]any `json:"variants,omitempty"`
	Segments       int              `json:"segments,omitempty"`
	UpdatedAt      string           `json:"updatedAt,omitempty"`
	PublicToken    string           `json:"publicToken,omitempty"`
	Playback       Playback         `json:"playback,omitempty"`
}

type Metrics struct {
	CDNBytes       int64   `json:"cdnBytes"`
	P2PBytes       int64   `json:"p2pBytes"`
	OffloadPercent float64 `json:"offloadPercent"`
	HashFailures   int64   `json:"hashFailures"`
	Fallbacks      int64   `json:"fallbacks"`
	Rebuffers      int64   `json:"rebuffers"`
	PeakPeers      int64   `json:"peakPeers"`
	Sessions       int64   `json:"sessions"`
}

type CreateStreamInput struct {
	URL   string `json:"url"`
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Live  *bool  `json:"live,omitempty"`
}

type TelemetryInput struct {
	SessionID    string `json:"sessionId"`
	StreamID     string `json:"streamId"`
	CDNBytes     int64  `json:"cdnBytes,omitempty"`
	P2PBytes     int64  `json:"p2pBytes,omitempty"`
	HashFailures int64  `json:"hashFailures,omitempty"`
	Fallbacks    int64  `json:"fallbacks,omitempty"`
	Rebuffers    int64  `json:"rebuffers,omitempty"`
	PeakPeers    int64  `json:"peakPeers,omitempty"`
}

func New(apiKey string) *Client {
	return NewWithOptions(apiKey, DefaultBaseURL, nil)
}

func NewWithOptions(apiKey, baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		APIKey:     apiKey,
		BaseURL:    strings.TrimRight(baseURL, "/"),
		HTTPClient: httpClient,
	}
}

func escape(value string) string {
	return url.PathEscape(value)
}

func (c *Client) Request(ctx context.Context, method, path string, body any, auth bool) (map[string]any, error) {
	if !strings.HasPrefix(c.APIKey, "key_") {
		return nil, fmt.Errorf("seedess: apiKey must be a server API key (key_…)")
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	if body != nil {
		req.Header.Set("content-type", "application/json")
	}
	if auth {
		req.Header.Set("authorization", "Bearer "+c.APIKey)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &data); err != nil {
			data = map[string]any{"error": map[string]any{"code": "invalid_json", "message": string(raw)}}
		}
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		code, message := "http_error", fmt.Sprintf("HTTP %d", res.StatusCode)
		if e, ok := data["error"].(map[string]any); ok {
			if v, ok := e["code"].(string); ok {
				code = v
			}
			if v, ok := e["message"].(string); ok {
				message = v
			}
		}
		return nil, &APIError{Status: res.StatusCode, Code: code, Message: message, Body: data}
	}
	return data, nil
}

func decode[T any](data map[string]any) (*T, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateStream(ctx context.Context, input CreateStreamInput) (*Stream, error) {
	data, err := c.Request(ctx, http.MethodPost, "/api/v1/streams", input, true)
	if err != nil {
		return nil, err
	}
	return decode[Stream](data)
}

func (c *Client) GetStream(ctx context.Context, id string) (*Stream, error) {
	data, err := c.Request(ctx, http.MethodGet, "/api/v1/streams/"+escape(id), nil, true)
	if err != nil {
		return nil, err
	}
	return decode[Stream](data)
}

func (c *Client) ListStreams(ctx context.Context) ([]Stream, error) {
	data, err := c.Request(ctx, http.MethodGet, "/api/v1/streams", nil, true)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(data["streams"])
	if err != nil {
		return nil, err
	}
	var streams []Stream
	if err := json.Unmarshal(raw, &streams); err != nil {
		return nil, err
	}
	return streams, nil
}

func (c *Client) RefreshStream(ctx context.Context, id string) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/v1/streams/"+escape(id)+"/refresh", nil, true)
}

func (c *Client) DisablePeers(ctx context.Context, id string) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/v1/streams/"+escape(id)+"/kill", nil, true)
}

func (c *Client) EnablePeers(ctx context.Context, id string) (map[string]any, error) {
	return c.Request(ctx, http.MethodPost, "/api/v1/streams/"+escape(id)+"/revive", nil, true)
}

func (c *Client) GetMetrics(ctx context.Context, id string) (*Metrics, error) {
	data, err := c.Request(ctx, http.MethodGet, "/api/v1/streams/"+escape(id)+"/metrics", nil, true)
	if err != nil {
		return nil, err
	}
	metrics, ok := data["metrics"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("seedess: metrics missing from response")
	}
	return decode[Metrics](metrics)
}

func (c *Client) SendTelemetry(ctx context.Context, input TelemetryInput) error {
	_, err := c.Request(ctx, http.MethodPost, "/api/v1/telemetry", input, false)
	return err
}
