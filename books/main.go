package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	goisbn "github.com/abx123/go-isbn"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type LambdaError struct {
	Code          int    `json:"code"`
	PublicMessage string `json:"public_message"`
}

func main() {
	lambda.Start(get)
}

func get(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
	}

	if isbn, ok := request.PathParameters["isbn"]; ok {
		gi := goisbn.NewGoISBN(goisbn.DEFAULT_PROVIDERS)

		book, err := gi.Get(isbn)
		if err != nil {
			errObj := &LambdaError{
				Code:          http.StatusInternalServerError,
				PublicMessage: err.Error(),
			}
			if err.Error() == "book not found" {
				errObj.Code = http.StatusNotFound
				errBody, _ := json.Marshal(errObj)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusNotFound,
					Headers:    headers,
					Body:       string(errBody),
				}, nil
			}
			errBody, _ := json.Marshal(errObj)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Headers:    headers,
				Body:       string(errBody),
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
			Body:       formatResp(book),
		}, nil
	} else {
		errObj := &LambdaError{
			Code:          http.StatusBadRequest,
			PublicMessage: fmt.Errorf("missing ISBN").Error(),
		}
		errBody, _ := json.Marshal(errObj)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
			Body:       string(errBody),
		}, nil
	}
}

func formatResp(input interface{}) string {
	bytesBuffer := new(bytes.Buffer)
	json.NewEncoder(bytesBuffer).Encode(input)
	responseBytes := bytesBuffer.Bytes()

	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, responseBytes, "", "  ")
	if error != nil {
		log.Println("JSON parse error: ", error)
	}
	formattedResp := prettyJSON.String()
	return formattedResp
}
