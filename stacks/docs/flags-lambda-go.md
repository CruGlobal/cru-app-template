Use [`github.com/CruGlobal/cru-flags`](https://github.com/CruGlobal/cru-flags)
(`go get github.com/CruGlobal/cru-flags`). Lambda freezes the function between
invocations, so a background refresh would not run: refresh from the handler
instead. Put this in `cmd/handler/flags.go`:

```go
package main

import (
	"context"
	"time"

	cruflags "github.com/CruGlobal/cru-flags"
)

// flags refreshes on demand, from the handler, at most once every 30 seconds.
var flags = cruflags.New(cruflags.Options{RefreshMode: cruflags.OnDemand})

// warmFlags waits for the first fetch during init, so the first invocation
// sees real flags. Settled or not, it carries on after 3 seconds.
func warmFlags() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	flags.Ready(ctx)
}
```

Then:

- call `warmFlags()` in `main`, before `lambda.Start`;
- call `flags.Refresh(ctx)` first thing in `handle`. It makes at most one
  conditional request every 30 seconds, and none while the flags are fresh;
- read a flag with `flags.Enabled("new_checkout")`. A read is a memory lookup
  that never fails.

Without `CRU_FLAGS_URL` (on your machine, in CI and in tests) every flag reads
off and nothing is fetched.
