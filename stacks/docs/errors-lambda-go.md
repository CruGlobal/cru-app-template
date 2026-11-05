Use [`github.com/rollbar/rollbar-go`](https://github.com/rollbar/rollbar-go)
(`go get github.com/rollbar/rollbar-go`). Put this in `cmd/handler/errors.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/rollbar/rollbar-go"
)

// setupErrorReporting points the Rollbar SDK at Flightdeck. The variables are
// absent locally and in CI, which is normal: then nothing is reported.
func setupErrorReporting() {
	token, endpoint := os.Getenv("ROLLBAR_ACCESS_TOKEN"), os.Getenv("ROLLBAR_ENDPOINT")
	// The Terraform module sets ROLLBAR_ENDPOINT on the function, but the token
	// only arrives through the secrets wrapper. An endpoint with no token means
	// the wrapper didn't run, or found no token, so say so loudly.
	if endpoint != "" && token == "" {
		slog.Error("ROLLBAR_ENDPOINT is set but ROLLBAR_ACCESS_TOKEN is not; " +
			"the function may be running without its secrets")
	}
	if token == "" || endpoint == "" {
		rollbar.SetEnabled(false)
		return
	}
	rollbar.SetToken(token)
	rollbar.SetEndpoint(endpoint)
	rollbar.SetEnvironment(os.Getenv("ENVIRONMENT"))
	rollbar.SetCodeVersion(os.Getenv("GIT_SHA"))
	// The SDK's default HTTP client never times out.
	rollbar.SetHTTPClient(&http.Client{Timeout: 3 * time.Second})
}

// reportErrors runs handle, reports the error it returns or a panic, and
// sends what is queued before the call ends. Lambda freezes the function the
// moment the handler returns, so a report still in the queue would not be sent.
// If you change handle's event or result type, change them here too.
func reportErrors(ctx context.Context, event json.RawMessage) (resp response, err error) {
	defer func() {
		p := recover()
		switch {
		case p != nil:
			rollbar.LogPanic(p, false)
		case err != nil:
			rollbar.Error(err)
		}
		flushErrors(3 * time.Second)
		if p != nil {
			panic(p)
		}
	}()
	return handle(ctx, event)
}

// flushErrors waits for queued reports to send, but never longer than limit.
// The SDK's own Wait has no time limit, and it retries a send that times out.
func flushErrors(limit time.Duration) {
	done := make(chan struct{})
	go func() {
		rollbar.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(limit):
	}
}
```

Then, in `cmd/handler/main.go`:

- call `setupErrorReporting()` before `lambda.Start`. The secrets wrapper has
  already put `ROLLBAR_ACCESS_TOKEN` in the environment by then. If it hasn't,
  the function logs an error at start, because then it has none of its
  secrets;
- start `reportErrors` in place of `handle`:
  `lambda.Start(ddlambda.WrapFunction(reportErrors, nil))`. Keep the Datadog
  wrapper on the outside, so its trace records the error too.

Report an error you handle, and carry on, with `rollbar.Error(err)`.
`reportErrors` sends it before the call ends.

Don't call `rollbar.Close()` in a Lambda function. Lambda reuses the process for
later calls, and `Close` stops the SDK for good, so every later report would be
lost. Don't use the SDK's `rollbar.LambdaWrapper` either: it waits with no time
limit, and it reports a panic but not a returned error. Keep the flush limit
well under the function's timeout.
