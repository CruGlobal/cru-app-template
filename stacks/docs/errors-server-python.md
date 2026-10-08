Use the [`rollbar`](https://pypi.org/project/rollbar/) package: add `rollbar`
to `requirements.txt`. Put this in `app/errors.py`:

```python
"""Error reporting to Flightdeck, through the Rollbar SDK."""
import os

import rollbar
import rollbar.contrib.flask

_TOKEN = os.environ.get("ROLLBAR_ACCESS_TOKEN")
_ENDPOINT = os.environ.get("ROLLBAR_ENDPOINT")

# The variables are absent locally and in CI, which is normal: then nothing is
# reported.
REPORTING = bool(_TOKEN and _ENDPOINT)


def init_error_reporting(app=None, *, job=False):
    """Pass the Flask app in a web server. Pass job=True in a one-off job or
    scheduled task, so each report is sent before the call returns and the job
    can exit right away."""
    if not REPORTING:
        return
    settings = {
        "environment": os.environ.get("ENVIRONMENT"),
        # The SDK adds "item/" to the endpoint itself, so pass the base.
        "endpoint": _ENDPOINT.removesuffix("item/"),
        "code_version": os.environ.get("GIT_SHA") or None,
    }
    if job:
        settings.update(handler="blocking", timeout=3)
    if app is not None:
        rollbar.contrib.flask.init(app, _TOKEN, **settings)
    else:
        rollbar.init(_TOKEN, **settings)


def report_exception():
    """Report the exception being handled. Call it inside an except block."""
    if REPORTING:
        rollbar.report_exc_info()
```

Call it right after the app is created in `app/main.py`. Errors raised while
Flask serves a request are then reported for you:

```python
app = Flask(__name__)
init_error_reporting(app)
```

Report an error you catch by calling `report_exception()` inside the `except`
block. A one-off job or scheduled task calls `init_error_reporting(job=True)`
instead. Don't call `rollbar.wait()`: it raises when reporting is off.

Watch the endpoint: `ROLLBAR_ENDPOINT` ends in `item/`, and this SDK adds
`item/` itself. Pass it whole and every report goes to a URL that doesn't exist,
with no error you would notice. The helper strips it for you.
