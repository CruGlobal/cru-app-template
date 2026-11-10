Use stock Flipper: the `flipper` and `flipper-active_support_cache_store` gems,
wired by the canonical initializer in the
[pipeline guide's Feature flags section](https://github.com/CruGlobal/cru-deploy/blob/main/docs/pipeline-v2.md#feature-flags).
Copy it exactly: the `Failsafe` adapter and the Marshal-safe read wrapper are
load-bearing. Then read a flag with `Flipper.enabled?(:new_checkout)`.

Without `CRU_FLAGS_URL` (on your machine, in CI and in tests) Flipper stays on
its in-memory adapter, where every flag reads off.
