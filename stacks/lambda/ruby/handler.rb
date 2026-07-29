# Minimal Cru Lambda handler — invoked by an event (API Gateway, SQS, an
# EventBridge schedule, etc.), not a web server with a port.
#
# Grow it into whatever the app needs: read `event`, do work, return a result.
# Keep the `handler(event:, context:)` signature and a fast, successful return
# — that's what the platform invokes and judges health by.
#
# DataDog: this handler is wrapped in code with Datadog::Lambda.wrap (the
# DataDog-documented Ruby approach — the gem ships no clean module-path CMD
# like Node/Python do), so it works with the standard AWS Ruby CMD
# ["handler.handler"] in the Dockerfile. Datadog::Lambda.wrap captures traces
# and metrics around the actual work. Remove the wrap and the inner block to
# run plain (un-instrumented).
require "json"
require "datadog/lambda"

Datadog::Lambda.configure_apm {}

def handler(event:, context:)
  Datadog::Lambda.wrap(event, context) do
    {
      statusCode: 200,
      body: { status: "ok" }.to_json,
    }
  end
end
