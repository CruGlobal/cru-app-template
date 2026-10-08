Use the [`rollbar`](https://pypi.org/project/rollbar/) package: add `rollbar`
to `requirements.txt`. Set it up at the top of `handler.py`, and report in the
handler:

```python
"""Error reporting to Flightdeck, through the Rollbar SDK."""
import os

import rollbar

_token = os.environ.get("ROLLBAR_ACCESS_TOKEN")
_endpoint = os.environ.get("ROLLBAR_ENDPOINT")

# Absent locally and in CI, which is normal: then nothing is reported.
REPORTING = bool(_token and _endpoint)
if REPORTING:
    rollbar.init(
        _token,
        environment=os.environ.get("ENVIRONMENT"),
        # The SDK adds "item/" to the endpoint itself, so pass the base.
        endpoint=_endpoint.removesuffix("item/"),
        code_version=os.environ.get("GIT_SHA") or None,
        # Send each report before the call returns, with a time limit: Lambda
        # freezes the function as soon as the handler returns.
        handler="blocking",
        timeout=3,
    )


def handler(event, context):
    try:
        # ... the function's work ...
        return {"statusCode": 200, "body": '{"status":"ok"}'}
    except Exception:
        if REPORTING:
            rollbar.report_exc_info()
        raise
```

Don't use the SDK's `@rollbar.lambda_function` decorator, and don't call
`rollbar.wait()`: both wait with no time limit, and `wait()` raises when
reporting is off.

Watch the endpoint: `ROLLBAR_ENDPOINT` ends in `item/`, and this SDK adds
`item/` itself. Pass it whole and every report goes to a URL that doesn't exist,
with no error you would notice. The code above strips it.
