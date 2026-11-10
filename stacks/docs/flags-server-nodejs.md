Use [`@cruglobal/flags`](https://github.com/CruGlobal/cru-flags)
(`npm install @cruglobal/flags`), on the server only: in a browser bundle every
flag reads off. In `src/index.ts`, wait for the first fetch before
`server.listen`, so the first requests after a deploy see real flags:

```ts
import { flags } from "@cruglobal/flags";

// Resolves once the first fetch settles, successfully or not.
await flags.ready();
```

Then read a flag anywhere on the server with `flags.enabled("new_checkout")`.
A timer refreshes the flags every 30 seconds without keeping the process
alive, and a read is a memory lookup that never throws. Without `CRU_FLAGS_URL`
(on your machine, in CI and in tests) every flag reads off and nothing is
fetched. To test both sides of a flag, pass the code a
`new CruFlags({ url })` pointed at a local server.

On Cloud Run with CPU only during requests, that timer stalls between
requests. Set `CRU_FLAGS_REFRESH_MODE=on-demand` and `await flags.refresh()`
at the start of each request instead: it fetches at most once every 30 seconds.
