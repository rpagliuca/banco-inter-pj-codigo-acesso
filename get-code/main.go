package main

import (
	"os"

	"bytes"
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse
type Request events.APIGatewayProxyRequest

const AWS_REGION = "us-east-1"

var GMAIL_IMAP_PASSWORD = os.Getenv("GMAIL_IMAP_PASSWORD")
var GMAIL_IMAP_USER = os.Getenv("GMAIL_IMAP_USER")

type Payload struct {
	Item string `json:"item"`
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request Request) (Response, error) {

	var buf bytes.Buffer

	body, err := json.Marshal(map[string]interface{}{
		"code": fetchCodeFromMailbox(GMAIL_IMAP_USER, GMAIL_IMAP_PASSWORD),
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
