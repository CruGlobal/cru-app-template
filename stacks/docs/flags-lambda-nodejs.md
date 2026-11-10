Use [`@cruglobal/flags`](https://github.com/CruGlobal/cru-flags)
(`npm install @cruglobal/flags`). Lambda freezes the function between
invocations, so a background timer would not run: refresh from the handler
instead. This stack bundles to CommonJS, which has no top-level `await`, so
the first fetch happens on the first invocation too. In `src/index.ts`:

```ts
import { CruFlags } from "@cruglobal/flags";

// Refreshes on demand, from the handler, at most once every 30 seconds.
const flags = new CruFlags({ refreshMode: "on-demand" });

export const handler = async (event: any): Promise<{ statusCode: number; body: string }> => {
  // Waits for the first fetch on a cold start; after that it is a no-op until
  // the flags are 30 seconds old, then one conditional request.
  await flags.refresh();

  if (flags.enabled("new_checkout")) {
    // ...
  }
  return { statusCode: 200, body: JSON.stringify({ status: "ok" }) };
};
```

A read is a memory lookup that never throws. Without `CRU_FLAGS_URL` (on your
machine, in CI and in tests) every flag reads off and nothing is fetched.
