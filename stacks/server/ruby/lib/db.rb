# frozen_string_literal: true

# Postgres connection helper for Cloud SQL. Nothing connects until you call
# Db.connect, so an app with no database is unaffected.
#
# Modes, chosen from environment variables:
#   password: DATABASE_URL or DATABASE_PASSWORD is set. Plain pg connection.
#   iam:      DATABASE_HOST, DATABASE_NAME and DATABASE_USER are set. A Google
#             login token is the password, sent over TCP with sslmode=require.
#   off:      anything else. Db.connect raises.
#
# There is no official Cloud SQL connector for Ruby, so IAM mode fetches the
# token itself with googleauth (application default credentials).
require "pg"

module Db
  IAM_VARS = %w[DATABASE_HOST DATABASE_NAME DATABASE_USER].freeze
  TOKEN_SCOPE = "https://www.googleapis.com/auth/sqlservice.login"
  TOKEN_LOCK = Mutex.new

  module_function

  def mode(env = ENV)
    return :password if present?(env["DATABASE_URL"]) || present?(env["DATABASE_PASSWORD"])
    return :iam if IAM_VARS.all? { |name| present?(env[name]) }

    :off
  end

  # Opens a new PG::Connection. The caller closes it, or use with_connection.
  def connect
    case mode
    when :password then PG.connect(password_params)
    when :iam then PG.connect(iam_params)
    else
      missing = IAM_VARS.reject { |name| present?(ENV[name]) }
      raise "Database is not configured. Set DATABASE_URL or DATABASE_PASSWORD, " \
            "or set #{missing.join(", ")}."
    end
  end

  def with_connection
    conn = connect
    yield conn
  ensure
    conn&.close
  end

  # Tokens last about an hour, so a fresh one is fetched for every new
  # connection. Never cache the result.
  def fresh_token
    require "googleauth"
    TOKEN_LOCK.synchronize do
      @credentials ||= Google::Auth.get_application_default(TOKEN_SCOPE)
      @credentials.fetch_access_token!
      @credentials.access_token
    end
  end

  def password_params
    # libpq reads PGSSLMODE on its own when it is not in the URL.
    return ENV["DATABASE_URL"] if present?(ENV["DATABASE_URL"])

    {
      host: ENV["DATABASE_HOST"],
      dbname: ENV["DATABASE_NAME"],
      user: ENV["DATABASE_USER"],
      password: ENV["DATABASE_PASSWORD"]
    }.compact
  end

  def iam_params
    # DATABASE_USER is the IAM database user exactly as Terraform sets it.
    {
      host: ENV["DATABASE_HOST"],
      dbname: ENV["DATABASE_NAME"],
      user: ENV["DATABASE_USER"],
      password: fresh_token,
      sslmode: "require"
    }
  end

  def present?(value)
    !value.to_s.empty?
  end
end
