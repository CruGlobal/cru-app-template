# frozen_string_literal: true

require "json"
# CRU:IAP
require "cru_iap"

# One JSON line per event on stdout; Cloud Run and Datadog read `severity`.
module Log
  def self.warn(message, **fields)
    $stdout.puts(JSON.generate(severity: "WARNING", message: message, **fields))
  end
end
CruIap.logger = Log
# /CRU:IAP
$stdout.sync = true

module App
  # CRU:IAP
  # IAP sign-in gate: 401 for all but /up.
  def self.authenticate(env)
    path = env["PATH_INFO"]
    # /up is the health probe: it calls the container directly, with no IAP assertion.
    return nil if path == "/up"

    result = CruIap.dev_bypass || CruIap::TokenVerifier.from_request(env)
    return result.email if result.ok?

    Log.warn("iap_rejected", reason: result.reason, path: path)
    false
  end

  # /CRU:IAP
  def self.call(env)
    # CRU:IAP
    email = authenticate(env)
    return [401, { "content-type" => "text/plain; charset=utf-8" }, ["Unauthorized"]] if email == false

    # /CRU:IAP
    case env["PATH_INFO"]
    # Health check: keep a 200 here or deploys are marked unhealthy.
    when "/up"
      [200, { "content-type" => "application/json" }, [{ status: "ok" }.to_json]]
    else
      # CRU:IAP
      [200, { "content-type" => "text/plain; charset=utf-8" }, ["Hello, #{email} 👋"]]
      # /CRU:IAP
      # CRU:ECS [200, { "content-type" => "text/plain; charset=utf-8" }, ["Hello 👋"]]
    end
  end
end
