"""Minimal Cru Lambda handler — invoked by an event (API Gateway, SQS, an
EventBridge schedule, etc.), not a web server with a port. The container's
DataDog wrapper (CMD datadog_lambda.handler.handler) calls this function via
the DD_LAMBDA_HANDLER env var set in the Dockerfile (handler.handler).

Grow it into whatever the app needs: read `event`, do work, return a result.
Keep the `handler(event, context)` signature and a fast, successful return —
that's what the platform invokes and judges health by.
"""


def handler(event, context):
    return {
        "statusCode": 200,
        "body": '{"status":"ok"}',
    }
