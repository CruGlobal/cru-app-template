require "minitest/autorun"
require "rack"
require "rack/mock_request"

class AppTest < Minitest::Test
  APP = Rack::Builder.parse_file(File.expand_path("../config.ru", __dir__))

  def setup
    @saved = ENV.to_h.slice("IAP_AUDIENCE", "CRU_IAP_DEV_BYPASS_EMAIL")
    ENV.delete("IAP_AUDIENCE")
    ENV.delete("CRU_IAP_DEV_BYPASS_EMAIL")
  end

  def teardown
    ENV.delete("IAP_AUDIENCE")
    ENV.delete("CRU_IAP_DEV_BYPASS_EMAIL")
    ENV.update(@saved)
  end

  def get(path) = Rack::MockRequest.new(APP).get(path)

  def test_up_is_always_open
    assert_equal 200, get("/up").status
  end

  def test_everything_else_needs_an_assertion_or_the_dev_bypass
    assert_equal 401, get("/").status
    assert_equal 401, get("/health").status
  end

  def test_dev_bypass_names_the_user
    ENV["CRU_IAP_DEV_BYPASS_EMAIL"] = "dev@example.com"
    response = get("/")
    assert_equal 200, response.status
    assert_equal "Hello, dev@example.com 👋", response.body
  end

  def test_iap_audience_ignores_the_dev_bypass
    ENV["IAP_AUDIENCE"] = "/projects/1/global/backendServices/2"
    ENV["CRU_IAP_DEV_BYPASS_EMAIL"] = "dev@example.com"
    assert_equal 401, get("/").status
    assert_equal 200, get("/up").status
  end
end
