# Seedess Go SDK

Official Seedess SDK for Go.

**Server-side only.** The SDK uses a secret API key (`key_…`) to create and control streams.
Never ship it in a browser or a mobile app. Playback uses the separate `@seedess/*-p2p`
packages and a per-stream public token (`pk_…`).

## Install

```bash
go get github.com/Seedess/go-sdk
```

## Use

```go
client := seedess.New(os.Getenv("SEEDESS_API_KEY"))

live := true
stream, err := client.CreateStream(ctx, seedess.CreateStreamInput{
    URL:  "https://origin.example/master.m3u8",
    ID:   "demo",
    Live: &live,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(stream.PublicToken)
fmt.Println(stream.Playback.Master)
```

Pass only the public token and stream id to the page.

## Methods

| Method | HTTP | Purpose |
| --- | --- | --- |
| `CreateStream(ctx, input)` | `POST /api/v1/streams` | Ingest a manifest, mint the public token. |
| `GetStream(ctx, id)` | `GET /api/v1/streams/:id` | Status, variants, segment count. |
| `ListStreams(ctx)` | `GET /api/v1/streams` | All streams in the tenant. |
| `RefreshStream(ctx, id)` | `POST /api/v1/streams/:id/refresh` | Pull new live segments. |
| `DisablePeers(ctx, id)` | `POST /api/v1/streams/:id/kill` | CDN only; playback continues. |
| `EnablePeers(ctx, id)` | `POST /api/v1/streams/:id/revive` | Re-enable peer delivery. |
| `GetMetrics(ctx, id)` | `GET /api/v1/streams/:id/metrics` | Offload and delivery counters. |
| `SendTelemetry(ctx, input)` | `POST /api/v1/telemetry` | One session batch (public, rate limited). |

`Request(ctx, method, path, body, auth)` is available for routes without a helper.

## Errors

```go
var apiErr *seedess.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.Status, apiErr.Code, apiErr.Message)
}
```

## Links

- Docs: https://seedess.com/docs/
- API: https://seedess.com/docs/api/
- Libraries: https://seedess.com/docs/libraries/

## License

MIT.
