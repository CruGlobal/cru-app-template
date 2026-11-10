Use the [`cru-flags`](https://github.com/CruGlobal/cru-flags) gem (add
`gem "cru-flags"` to the `Gemfile`). This Rack stack loads gems by hand, so in
`config.ru`, before `run`, load it and wait for the first fetch, so the first
requests after a deploy see real flags:

```ruby
require "cru_flags"

# Settled or not, carry on after 3 seconds: until a fetch lands, every flag
# reads off.
CruFlags.ready(timeout: 3.0)
```

Then read a flag anywhere with `CruFlags.enabled?("new_checkout")`. A
background thread refreshes the flags every 30 seconds, and a read is a memory
lookup that never raises. Without `CRU_FLAGS_URL` (on your machine, in CI and
in tests) every flag reads off and nothing is fetched.

If the app grows into Rails, the gem's Railtie points Flipper at the flag
service by itself: read flags with `Flipper.enabled?(:new_checkout)`, and add
no Flipper initializer of your own, because one would replace the gem's. To
turn flags on in Rails tests, configure Flipper's memory adapter in the test
environment.
