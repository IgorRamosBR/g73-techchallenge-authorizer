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
	lambda.Start(LambdaHandler)
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
		log.Printf("failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func createUserHandler(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		log.Printf("failed to bind user: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := saveUserToDynamoDB(c, user); err != nil {
		log.Printf("failed to save user, error: %v", err)
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
		log.Printf("failed to load default config, error: %v", err)
		return err
	}
	client := dynamodb.NewFromConfig(cfg)

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		log.Printf("failed to create item, error: %v", err)
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	}

	_, err = client.PutItem(ctx, input)
	if err != nil {
		log.Printf("failed toto put item, error: %v", err)
		return err
	}

	return nil
}
