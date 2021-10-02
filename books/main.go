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
			if err.Error() == "book not found" {
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusNotFound,
					Headers:    headers,
					Body:       err.Error(),
				}, err
			}
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Headers:    headers,
				Body:       err.Error(),
			}, err
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
			Body:       formatResp(book),
		}, nil
	} else {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
			Body:       "missing ISBN",
		}, fmt.Errorf("missing ISBN")
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
