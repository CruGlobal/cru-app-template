Use [`github.com/rollbar/rollbar-go`](https://github.com/rollbar/rollbar-go)
(`go get github.com/rollbar/rollbar-go`). Add this to `cmd/server/main.go`:

```go
// setupErrorReporting points the Rollbar SDK at Flightdeck. The variables are
// absent locally and in CI, which is normal: then nothing is reported.
func setupErrorReporting() {
	token, endpoint := os.Getenv("ROLLBAR_ACCESS_TOKEN"), os.Getenv("ROLLBAR_ENDPOINT")
	if token == "" || endpoint == "" {
		rollbar.SetEnabled(false)
		return
	}
	rollbar.SetToken(token)
	rollbar.SetEndpoint(endpoint)
	rollbar.SetEnvironment(os.Getenv("ENVIRONMENT"))
	rollbar.SetCodeVersion(os.Getenv("GIT_SHA"))
	// The SDK's default HTTP client never times out.
	rollbar.SetHTTPClient(&http.Client{Timeout: 5 * time.Second})
}

// reportPanics reports a panic in a request handler, then panics again so
// net/http still logs it and drops the connection.
func reportPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				if p != http.ErrAbortHandler {
					rollbar.LogPanic(p, true)
				}
				panic(p)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

Then, in `main`:

- call `setupErrorReporting()` first;
- wrap the routes, as in `Handler: reportPanics(routes())`;
- call `rollbar.Close()` before `main` returns or calls `os.Exit`. `Close` sends
  what is still queued, and the client above puts a time limit on each send. A
  deferred call does not run on `os.Exit`, so call it before.

Report an error you handle with `rollbar.Error(err)`. A job built into the same
image does the same: set up first, then `rollbar.Close()` before it exits.
