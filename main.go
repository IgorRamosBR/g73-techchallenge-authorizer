package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type User struct {
	CPF   string `json:"cpf" dynamodbv:"CPF"`
	Name  string `json:"name" dynamodbv:"Name"`
	Email string `json:"email" dynamodbv:"Email"`
}

type Response struct {
	IsAuthorized bool   `json:"isAuthorized"`
	UserId       int    `json:"id"`
	Message      string `json:"message"`
}

var (
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
)

var ginLambda *ginadapter.GinLambda

func init() {
	log.Printf("Gin cold start")
	router := gin.Default()

	router.GET("/authorize", authorizeUserHandler)
	router.POST("/user", createUserHandler)

	ginLambda = ginadapter.New(router)
}

func LambdaHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(ginLambda)
}

func authorizeUserHandler(c *gin.Context) {
	var requestBody struct {
		CPF string `json:"cpf"`
	}
	err := c.BindJSON(&requestBody)
	if err != nil {
		log.Printf("Bad request: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User unauthorized"})
		return
	}

	item, err := getUserFromDynamoDB(c, requestBody.CPF)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func createUserHandler(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := saveUserToDynamoDB(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.Status(http.StatusCreated)
}

func getUserFromDynamoDB(ctx context.Context, cpf string) (User, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return User{}, err
	}

	client := dynamodb.NewFromConfig(cfg)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"CPF": &types.AttributeValueMemberS{Value: cpf},
		},
	}

	result, err := client.GetItem(ctx, input)
	if err != nil {
		return User{}, err
	}

	if len(result.Item) == 0 {
		return User{}, fmt.Errorf("user with CPF %s not found", cpf)
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Item, user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func saveUserToDynamoDB(ctx context.Context, user User) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := dynamodb.NewFromConfig(cfg)

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}

	_, err = client.PutItem(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

// func main() {
// 	// Connect to the RDS database
// 	db, err := sql.Open("postgres", "host=g73-techchallenge-db.cxokeewukuer.us-east-1.rds.amazonaws.com port=5432 user=g73_admin_user password=UV6RetyeibtF dbname=mydb sslmode=disable")
// 	if err != nil {
// 		log.Fatalf("Error connecting to database: %v", err)
// 	}
// 	defer db.Close()

// 	handler := func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
// 		var requestBody struct {
// 			CPF string `json:"cpf"`
// 		}
// 		if err := json.Unmarshal([]byte(request.Body), &requestBody); err != nil {
// 			log.Printf("Bad request: %v", err)
// 			return events.APIGatewayProxyResponse{StatusCode: 404}, nil
// 		}

// 		var user User
// 		err := db.QueryRow("SELECT id, name, cpf FROM users WHERE cpf = ?", requestBody.CPF).Scan(&user.ID, &user.Name, &user.CPF)
// 		if err != nil {
// 			if err == sql.ErrNoRows {
// 				return events.APIGatewayProxyResponse{StatusCode: 403, Body: "Unauthorized"}, nil
// 			}
// 			return events.APIGatewayProxyResponse{StatusCode: 500}, fmt.Errorf("error querying database: %v", err)
// 		}

// 		response := Response{IsAuthorized: true, UserId: user.ID, Message: "User found"}
// 		responseBody, err := json.Marshal(response)
// 		if err != nil {
// 			log.Printf("Error marshalling response body: %v", err)
// 			return events.APIGatewayProxyResponse{StatusCode: 500}, nil
// 		}

// 		return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(responseBody)}, nil
// 	}

// 	lambda.Start(handler)
// }
