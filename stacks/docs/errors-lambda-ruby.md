Use the [`rollbar`](https://rubygems.org/gems/rollbar) gem: add
`gem "rollbar"` to the Gemfile. Put this in `errors.rb`, next to `handler.rb`:

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

Then report in the handler, inside the Datadog wrap:

```ruby
require_relative "errors"

def handler(event:, context:)
  Datadog::Lambda.wrap(event, context) do
    # ... the function's work ...
    { statusCode: 200, body: { status: "ok" }.to_json }
  rescue => e
    # Sent before this returns: the gem reports synchronously unless use_async
    # is on, and Lambda freezes the function once the handler returns.
    Rollbar.error(e)
    raise
  end
end
```

Leave `use_async` off in a Lambda function, for the reason in the comment.
