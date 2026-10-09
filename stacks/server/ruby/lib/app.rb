# frozen_string_literal: true

require "json"
require "cru_iap"

# One JSON line per event on stdout; Cloud Run and Datadog read `severity`.
module Log
  def self.warn(message, **fields)
    $stdout.puts(JSON.generate(severity: "WARNING", message: message, **fields))
  end
end
$stdout.sync = true
CruIap.logger = Log

module App
  # IAP sign-in gate: 401 for all but /up. ECS has no IAP in front, so an ECS app
  # must remove or replace it.
  def self.authenticate(env)
    path = env["PATH_INFO"]
    # /up is the health probe: it calls the container directly, with no IAP assertion.
    return nil if path == "/up"

    result = CruIap.dev_bypass || CruIap::TokenVerifier.from_request(env)
    return result.email if result.ok?

    Log.warn("iap_rejected", reason: result.reason, path: path)
    false
  end

  def self.call(env)
    email = authenticate(env)
    return [401, { "content-type" => "text/plain; charset=utf-8" }, ["Unauthorized"]] if email == false

    case env["PATH_INFO"]
    # Health check: keep a 200 here or deploys are marked unhealthy.
    when "/health", "/up"
      [200, { "content-type" => "application/json" }, [{ status: "ok" }.to_json]]
    else
      [200, { "content-type" => "text/plain; charset=utf-8" }, ["Hello, #{email} 👋"]]
    end
  end
end
