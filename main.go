package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	CPF  string `json:"cpf"`
}

type Response struct {
	UserId       int    `json:"userId"`
	IsAuthorized bool   `json:"isAuthorized"`
	Message      string `json:"message"`
}

func main() {
	db, err := sql.Open("postgres", getSQLConnectionURL())
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	handler := func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var requestBody struct {
			CPF string `json:"cpf"`
		}
		if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
			log.Printf("Error unmarshalling request body: %v", err)
			return events.APIGatewayProxyResponse{StatusCode: 400}, nil
		}

		var user User
		err := db.QueryRow("SELECT id, name, cpf FROM public.customers WHERE cpf = $1", requestBody.CPF).Scan(&user.ID, &user.Name, &user.CPF)
		if err != nil {
			if err == sql.ErrNoRows {
				return events.APIGatewayProxyResponse{StatusCode: 403, Body: "Unauthorized"}, nil
			}
			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error querying database: %v", err)
		}

		response := Response{UserId: user.ID, IsAuthorized: true, Message: fmt.Sprintf("User [%d] is authorized.", user.ID)}
		responseBody, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshalling response body: %v", err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, nil
		}

		return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(responseBody)}, nil
	}

	lambda.Start(handler)
}

func getSQLConnectionURL() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dbConnecction := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, port, user, password, dbName)
	return dbConnecction
}
