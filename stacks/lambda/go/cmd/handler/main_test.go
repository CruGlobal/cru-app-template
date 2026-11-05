package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/lambda"
)

// TestHandle invokes the handler the way the Lambda runtime does, through
// lambda.NewHandler, so it also checks the handler's signature and the JSON it
// returns.
func TestHandle(t *testing.T) {
	got, err := lambda.NewHandler(handle).Invoke(t.Context(), []byte(`{}`))
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	want := `{"statusCode":200,"body":"{\"status\":\"ok\"}"}`
	if string(got) != want {
		t.Errorf("Invoke() = %s, want %s", got, want)
	}
}
