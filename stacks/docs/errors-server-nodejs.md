Use the [`rollbar`](https://www.npmjs.com/package/rollbar) package
(`npm install rollbar`). Put this in `src/errors.ts`, and import it at the top
of `src/index.ts` so it runs before anything else:

```ts
// Error reporting to Flightdeck, through the Rollbar SDK.
import Rollbar from "rollbar";

const { ROLLBAR_ACCESS_TOKEN, ROLLBAR_ENDPOINT, ENVIRONMENT, GIT_SHA } = process.env;

// The variables are absent locally and in CI, which is normal: then this is
// null and nothing is reported.
export const rollbar =
  ROLLBAR_ACCESS_TOKEN && ROLLBAR_ENDPOINT
    ? new Rollbar({
        accessToken: ROLLBAR_ACCESS_TOKEN,
        endpoint: ROLLBAR_ENDPOINT,
        environment: ENVIRONMENT,
        codeVersion: GIT_SHA || undefined,
      })
    : null;

// Waits for queued reports to send, but never longer than `ms`. The SDK's own
// wait has no time limit.
export function flushErrors(ms = 5000): Promise<void> {
  const notifier = rollbar;
  if (!notifier) return Promise.resolve();
  return new Promise((resolve) => {
    const timer = setTimeout(resolve, ms);
    notifier.wait(() => {
      clearTimeout(timer);
      resolve();
    });
  });
}

// Report a crash, then exit the way Node would have. The SDK's captureUncaught
// keeps a crashed process running unless told to exit, and its exit waits with
// no time limit, so this does both itself.
process.on("uncaughtException", (err) => {
  console.error(err);
  rollbar?.error(err);
  void flushErrors().finally(() => process.exit(1));
});
```

Report an error you catch with `rollbar?.error(err)`. To add details, pass a
plain object as the second argument. Don't pass the raw request, which carries
cookies and headers.

A one-off job or script built into the same image imports the same module and
calls `await flushErrors()` before it exits. Leave the SDK's `captureUncaught`
off: its listener keeps a crashed process running, and the handler above
already reports a crash and exits.
