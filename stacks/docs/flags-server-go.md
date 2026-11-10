Use [`github.com/CruGlobal/cru-flags`](https://github.com/CruGlobal/cru-flags)
(`go get github.com/CruGlobal/cru-flags`). In `run` in `cmd/server/main.go`,
wait for the first fetch before `srv.ListenAndServe`, so the first requests
after a deploy see real flags:

```go
// Settled or not, carry on after 3 seconds: until a fetch lands, every flag
// reads off.
readyCtx, cancelReady := context.WithTimeout(ctx, 3*time.Second)
cruflags.Ready(readyCtx)
cancelReady()
```

Then read a flag anywhere:

```go
import cruflags "github.com/CruGlobal/cru-flags"

if cruflags.Enabled("new_checkout") {
	// ...
}
```

A goroutine refreshes the flags every 30 seconds, and a read is a memory
lookup that never fails. Without `CRU_FLAGS_URL` (on your machine, in CI and in
tests) every flag reads off and nothing is fetched. To test both sides of a
flag, pass the code a `*cruflags.Client` built with
`cruflags.New(cruflags.Options{URL: srv.URL})` against an `httptest` server.

On Cloud Run with CPU only during requests, that goroutine stalls between
requests. Set `CRU_FLAGS_REFRESH_MODE=on-demand` and call
`cruflags.Refresh(r.Context())` in middleware instead: it fetches at most once
every 30 seconds.
