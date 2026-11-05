// Command handler is a minimal Cru Lambda function: it is invoked by an event
// (API Gateway, SQS, an EventBridge schedule, and so on), not a web server
// with a port. The Dockerfile builds it as the bootstrap binary that Lambda's
// OS-only runtime runs.
//
// Grow it into whatever the app needs: read the event, do work, return a
// result. Keep a fast, successful return: that's what the platform invokes and
// judges health by.
package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	ddlambda "github.com/DataDog/dd-trace-go/contrib/aws/datadog-lambda-go/v2"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	// JSON lines on stdout. Lambda passes them to the Datadog extension, which
	// parses them into fields.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	// slog.SetDefault also routes the log package through slog, at INFO. The
	// Datadog wrapper writes its own JSON lines, errors included, with the log
	// package, so send those straight to stdout to keep their level.
	log.SetOutput(os.Stdout)

	// The Datadog wrapper sends enhanced metrics and a trace for every call to
	// the Datadog extension the Dockerfile copies into /opt. Keep it around
	// the handler.
	lambda.Start(ddlambda.WrapFunction(handle, nil))
}

// response is the shape API Gateway and function URLs expect. Other triggers
// ignore what the handler returns.
type response struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

// handle does the function's work for one event. Decode the event into the
// type your trigger sends: github.com/aws/aws-lambda-go/events has them all.
func handle(_ context.Context, _ json.RawMessage) (response, error) {
	return response{StatusCode: http.StatusOK, Body: `{"status":"ok"}`}, nil
}
