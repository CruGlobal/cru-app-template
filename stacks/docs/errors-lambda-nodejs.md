Use the [`rollbar`](https://www.npmjs.com/package/rollbar) package
(`npm install rollbar`). Put this in `src/errors.ts`:

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
```

Then report and flush in the handler, in `src/index.ts`:

```ts
import { flushErrors, rollbar } from "./errors.js";

export const handler = async (event: any) => {
  try {
    // ... the function's work ...
    return { statusCode: 200, body: JSON.stringify({ status: "ok" }) };
  } catch (err) {
    rollbar?.error(err as Error);
    // Lambda freezes the function the moment the handler returns, so a report
    // still in the queue would never be sent.
    await flushErrors();
    throw err;
  }
};
```

Don't use the SDK's `rollbar.lambdaHandler` wrapper: it waits for the send with
no time limit. Keep `NODE_OPTIONS=--enable-source-maps` in the Dockerfile, so
stack traces point at your TypeScript. Nothing is uploaded for Lambda.
