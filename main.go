package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

type User struct {
	CPF     string `json:"cpf" dynamodbv:"CPF"`
	Name    string `json:"name" dynamodbv:"Name"`
	Email   string `json:"email" dynamodbv:"Email"`
	Address string `json:"address" dynamodbv:"Address"`
	Phone   string `json:"phone" dynamodbv:"Phone"`
}

type Response struct {
	IsAuthorized bool   `json:"isAuthorized"`
	Message      string `json:"message"`
	User         User   `json:"user,omitempty"`
}

var ErrUserNotFound = errors.New("user not found")

var (
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
)

var ginLambda *ginadapter.GinLambda

func init() {
	log.Printf("Gin cold start")
	router := gin.Default()

	router.POST("/authorize", authorizeUserHandler)
	router.POST("/user", createUserHandler)
	router.PUT("/user/:cpf/clean", cleanUserDataHandler)

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user unauthorized"})
		return
	}

	user, err := getUserFromDynamoDB(c, requestBody.CPF)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response := Response{
				IsAuthorized: false,
				Message:      "user unauthorized",
			}
			log.Printf("User unauthorized: %s", requestBody.CPF)
			c.JSON(http.StatusUnauthorized, response)
			return
		}

		log.Printf("failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	response := Response{
		IsAuthorized: true,
		Message:      "user authorized",
		User:         user,
	}
	c.JSON(http.StatusOK, response)
}

func createUserHandler(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		log.Printf("failed to bind user: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := saveUserToDynamoDB(c, user); err != nil {
		log.Printf("failed to save user, error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.Status(http.StatusCreated)
}

func cleanUserDataHandler(c *gin.Context) {
	cpf := c.Param("cpf")
	if cpf == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cpf must not be empty"})
		return
	}

	err := cleanUserData(c, cpf)
	if err != nil {
		log.Printf("failed to deactivate user, error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate user"})
		return
	}

	c.Status(http.StatusNoContent)
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
		return User{}, ErrUserNotFound
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Item, &user)
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

func cleanUserData(ctx context.Context, cpf string) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("failed to load default config, error: %v", err)
		return err
	}
	client := dynamodb.NewFromConfig(cfg)

	key := map[string]types.AttributeValue{
		"CPF": &types.AttributeValueMemberS{Value: cpf},
	}

	update := expression.Remove(expression.Name("Name")).Remove(expression.Name("Email")).Remove(expression.Name("Phone")).Remove(expression.Name("Address"))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		log.Printf("failed to build expression, %v", err)
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableName),
		Key:                       key,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}

	_, err = client.UpdateItem(context.TODO(), input)
	if err != nil {
		log.Printf("failed to update user, %v", err)
		return err
	}

	return nil
}
