# Minimal Cru app — a tiny Rack app with a health check, so the container
# builds and deploys as-is. It boots in seconds with no asset pipeline. The
# routes are in lib/app.rb.
# CRU:IAP
# So is the IAP sign-in gate.
# /CRU:IAP
#
# Prefer full Rails? Run `rails new . --force` in this directory (after adding
# the `rails` gem and running `bundle install`), then update the Dockerfile's
# CMD to boot Rails (e.g. `./bin/rails server -p $PORT`). Rails' own health
# check lives at /up, which this starter already answers.
# CRU:IAP
# Carry the IAP gate over as the cru-iap README's Rails section shows, then
# delete lib/app.rb.
# /CRU:IAP
#
# Only development runs read .env.development; variables already set win.
if ENV.fetch("RACK_ENV", "development") == "development" && File.exist?(".env.development")
  File.foreach(".env.development") do |line|
    key, sep, value = line.strip.partition("=")
    key, value = key.strip, value.strip
    # Like Node's --env-file, drop one pair of quotes around the value.
    value = value[1..-2] if value.length >= 2 && %w[" '].include?(value[0]) && value[-1] == value[0]
    ENV[key] ||= value unless sep.empty? || key.start_with?("#")
  end
end

require_relative "lib/app"

# CRU:IAP
# Rack prefers X-Forwarded-Host over Host. Google's load balancer never sets
# it, so a value is always client-forged: drop it before anything reads it.
use CruIap::StripForwardedHost
# /CRU:IAP
run App
