# frozen_string_literal: true

# The app's routes. config.ru serves them; test/app_test.rb tests them.
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
  # The sign-in gate. On Cloud Run, Google IAP signs people in with Okta before
  # a request gets here, and the platform sets IAP_AUDIENCE: then every route
  # but /up needs IAP's signed assertion, and anything else is a 401. Without
  # IAP_AUDIENCE (ECS, or local dev) there is no gate;
  # CRU_IAP_DEV_BYPASS_EMAIL=you@cru.org gives you a signed-in email locally.
  # Sign out with /?gcp-iap-mode=CLEAR_LOGIN_COOKIE (CruIap.logout_url).
  # Returns the email (nil when there is no gate), or false to reject.
  def self.authenticate(env)
    path = env["PATH_INFO"]
    return nil if path == "/up" # the health check; the load balancer lets it skip IAP

    # Gate on deploy config, never on the header being absent.
    return CruIap.dev_bypass&.email if ENV.fetch("IAP_AUDIENCE", "").empty?

    result = CruIap::TokenVerifier.from_request(env)
    return result.email if result.ok?

    # Fail closed: never fall back to a dev identity here.
    Log.warn("iap_rejected", reason: result.reason, path: path)
    false
  end

  def self.call(env)
    email = authenticate(env)
    return [401, { "content-type" => "text/plain; charset=utf-8" }, ["Unauthorized"]] if email == false

    case env["PATH_INFO"]
    # Health check — the platform pings this to know the app is alive. Keep a
    # 200 here working or deploys will be marked unhealthy.
    when "/health", "/up"
      [200, { "content-type" => "application/json" }, [{ status: "ok" }.to_json]]
    else
      greeting = email ? "Hello, #{email} 👋" : "Hello from your Cru app 👋"
      [200, { "content-type" => "text/plain; charset=utf-8" }, [greeting]]
    end
  end
end
