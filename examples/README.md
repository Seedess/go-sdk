# Go SDK example

A complete server + browser sample:

- `server/main.go` — creates a stream with the Go SDK, then serves the page and `/api/seedess`.
- `server/public/index.html` — loads hls.js and `https://seedess.com/p2p.js`, then plays the
  stream with the P2P adapter.

## Run

```bash
cd sdks/go-sdk
SEEDESS_API_KEY=key_… go run ./examples/server
open http://localhost:8787
```

Optional: `SEEDESS_STREAM_URL`, `SEEDESS_STREAM_ID`, `SEEDESS_API_BASE`, `PORT`.

The first request to `/api/seedess` creates the stream. The API key stays in the server
process; the browser only sees the public token and playback URLs.
