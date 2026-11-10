Use [`cru-flags`](https://github.com/CruGlobal/cru-flags) (add `cru-flags` to
`requirements.txt`). In `app/main.py`, wait for the first fetch when the app
loads, so the first requests after a deploy see real flags:

```python
from cru_flags import flags

# Settled or not, carry on after 3 seconds: until a fetch lands, every flag
# reads off.
flags.ready(timeout=3.0)
```

Then read a flag anywhere with `flags.enabled("new_checkout")`. A daemon
thread refreshes the flags every 30 seconds, and a read is a memory lookup that
never raises. gunicorn imports the app in each worker, so each worker has its
own client. Without `CRU_FLAGS_URL` (on your machine, in CI and in tests) every
flag reads off and nothing is fetched. To test both sides of a flag, pass the
code a `cru_flags.Client(url=...)` pointed at a local server.
