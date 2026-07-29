# Minimal Cru app — a tiny Rack app with a health check, so the container
# builds and deploys as-is. It boots in seconds with no asset pipeline.
#
# Prefer full Rails? Run `rails new . --force` in this directory (after adding
# the `rails` gem and running `bundle install`), then update the Dockerfile's
# CMD to boot Rails (e.g. `./bin/rails server -p $PORT`). Rails' own health
# check lives at /up, which this starter already answers.
require "json"

run lambda { |env|
  case env["PATH_INFO"]
  # Health check — the platform pings this to know the app is alive. Keep a
  # 200 here working or deploys will be marked unhealthy.
  when "/health", "/up"
    [200, { "content-type" => "application/json" }, [{ status: "ok" }.to_json]]
  else
    [200, { "content-type" => "text/plain; charset=utf-8" }, ["Hello from your Cru app 👋"]]
  end
}
