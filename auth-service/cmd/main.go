package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
)

//	@title			Auth API
//	@version		1.0
//	@description	This is an auth api for an application.

// @BasePath	/api/v1
var ginLambda *ginadapter.GinLambda

// func init() {
// 	prod := os.Getenv("RELEASE_MODE")
// 	if prod == "true" {
// 		gin.SetMode(gin.ReleaseMode)
// 	}
// 	router := gin.Default()
// 	docs.SwaggerInfo.BasePath = "/api/v1"

// 	api := router.Group("/api/v1")
// 	//* Passing the router to all user(auth-service) routes.
// 	routes.UserRoute(api)

// 	//* Connecting to DB
// 	configs.ConnectDB()
// 	router.GET("/", controllers.BaseRoute)
// 	ginLambda = ginadapter.New(router)
// }

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
	// err := godotenv.Load("./.env")
	// if err != nil {
	// 	log.Println("Error loading environment")
	// 	return
	// }
	// prod := os.Getenv("RELEASE_MODE")
	// if prod == "true" {
	// 	gin.SetMode(gin.ReleaseMode)
	// }
	// router := gin.Default()
	// docs.SwaggerInfo.BasePath = "/api/v1"

	// api := router.Group("/api/v1")
	// //* Passing the router to all user(auth-service) routes.
	// routes.UserRoute(api)

	// //* Connecting to DB
	// configs.ConnectDB()
	// router.GET("/", controllers.BaseRoute)
	// // router.GET("/api/v1/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// router.Run("localhost:8080")
	// configs.CloseDB()
}
