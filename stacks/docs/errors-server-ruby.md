Use the [`rollbar`](https://rubygems.org/gems/rollbar) gem: add
`gem "rollbar"` to the Gemfile. Put this in `lib/errors.rb`:

```ruby
require "rollbar"

# Error reporting to Flightdeck, through the Rollbar SDK. The variables are
# absent locally and in CI, which is normal: then nothing is reported.
Rollbar.configure do |config|
  token = ENV["ROLLBAR_ACCESS_TOKEN"].to_s
  endpoint = ENV["ROLLBAR_ENDPOINT"].to_s
  config.enabled = !token.empty? && !endpoint.empty?
  config.access_token = token
  config.endpoint = endpoint unless endpoint.empty?
  config.environment = ENV["ENVIRONMENT"]
  config.code_version = ENV["GIT_SHA"] unless ENV["GIT_SHA"].to_s.empty?
end
```

Then add two lines to `config.ru`, before `run`. Errors raised while serving a
request are then reported for you:

```ruby
require_relative "lib/errors"
use Rollbar::Middleware::Rack
```

Report an error you rescue with `Rollbar.error(e)`. The gem sends each report
before the call returns (unless you turn on `use_async`), so a job or script has
nothing to flush before it exits.

If you move to Rails, `bin/rails generate rollbar` writes
`config/initializers/rollbar.rb`. Put the same settings there, including
`config.endpoint`.
