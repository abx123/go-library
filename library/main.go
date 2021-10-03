package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Book struct {
	Title               string      `json:"title"`
	PublishedYear       string      `json:"published_year"`
	Authors             []string    `json:"authors"`
	Description         string      `json:"description"`
	IndustryIdentifiers *Identifier `json:"industry_identifiers"`
	PageCount           int64       `json:"page_count"`
	Categories          []string    `json:"categories"`
	ImageLinks          *ImageLinks `json:"image_links"`
	Publisher           string      `json:"publisher"`
	Language            string      `json:"language"`
	Source              string      `json:"source"`
	UserID              string      `json:"user_id"`
	Status              int         `json:"status"`
}

type Identifier struct {
	ISBN   string `json:"isbn"`
	ISBN13 string `json:"isbn_13"`
}

type ImageLinks struct {
	SmallImageURL string `json:"small_image_url"`
	ImageURL      string `json:"image_url"`
	LargeImageURL string `json:"large_image_url"`
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
	req := Book{}
	err := json.Unmarshal([]byte(request.Body), &req)
	if (req.IndustryIdentifiers.ISBN13 == "" && req.IndustryIdentifiers.ISBN == "") || req.UserID == "" {
		err = fmt.Errorf("missing ISBN13 or UserID")
	}
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}
	client, _ := NewMongoClient()
	defer client.Disconnect(context.Background())
	collection := client.Database("library").Collection("books")
	res, err := upsertMongo(collection, &req)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    headers,
			Body:       err.Error(),
		}, nil
	}
	if res.MatchedCount == 1 {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusAccepted,
			Headers:    headers,
			Body:       "record updated.",
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers:    headers,
		Body:       "record created.",
	}, nil
}

func upsertMongo(collection *mongo.Collection, book *Book) (*mongo.UpdateResult, error) {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{
		"industryIdentifier.isbn":   book.IndustryIdentifiers.ISBN,
		"industryIdentifier.isbn13": book.IndustryIdentifiers.ISBN13,
		"userID":                    book.UserID}
	update := bson.M{
		"$set": bson.M{
			"title":              book.Title,
			"publishedYear":      book.PublishedYear,
			"authors":            book.Authors,
			"description":        book.Description,
			"industryIdentifier": book.IndustryIdentifiers,
			"pageCount":          book.PageCount,
			"categories":         book.Categories,
			"imageLinks":         book.ImageLinks,
			"publisher":          book.Publisher,
			"language":           book.Language,
			"source":             book.Source,
			"status":             book.Status,
			"userID":             book.UserID}}

	result, err := collection.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func NewMongoClient() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	return client, err
}
