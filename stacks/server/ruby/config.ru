# Minimal Cru app — a tiny Rack app with a health check, so the container
# builds and deploys as-is. It boots in seconds with no asset pipeline. The
# routes, and the IAP sign-in gate, are in lib/app.rb.
#
# Prefer full Rails? Run `rails new . --force` in this directory (after adding
# the `rails` gem and running `bundle install`), then update the Dockerfile's
# CMD to boot Rails (e.g. `./bin/rails server -p $PORT`). Rails' own health
# check lives at /up, which this starter already answers. Carry the IAP gate
# over as the cru-iap README's Rails section shows, then delete lib/app.rb.
# Local runs (puma defaults RACK_ENV to development; the Dockerfile sets
# production) read .env.development. Variables already set win.
if ENV.fetch("RACK_ENV", "development") == "development" && File.exist?(".env.development")
  File.foreach(".env.development") do |line|
    key, sep, value = line.strip.partition("=")
    ENV[key] ||= value unless sep.empty? || key.start_with?("#")
  end
end

require_relative "lib/app"

# Rack prefers X-Forwarded-Host over Host. Google's load balancer never sets
# it, so a value is always client-forged: drop it before anything reads it.
use CruIap::StripForwardedHost
run App
