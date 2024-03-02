package main

import (
	"auth-service/configs"
	controller "auth-service/controllers"
	docs "auth-service/docs"
	"auth-service/routes"
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Registration API
//	@version		1.0
//	@description	This is a registration api for an application.

// @BasePath	/api/
var ginLambda *ginadapter.GinLambda

func init() {

	prod := os.Getenv("RELEASE_MODE")
	if prod == "true" {
		gin.SetMode(gin.ReleaseMode)
	}
	godotenv.Load("../../.env")
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"

	api := router.Group("/api/v1")
	//* Passing the router to all user(auth-service) routes.
	routes.UserRoute(api)

	//* Connecting to DB
	configs.ConnectDB()
	router.GET("/", controller.BaseRoute)
	router.GET("/api/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	// router.Run(":8000")
	ginLambda = ginadapter.New(router)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
