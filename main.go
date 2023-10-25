package main

import (
	"gin-mongo-api/configs"
	"gin-mongo-api/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	configs.ConnectDB()
	routes.UserRoute(router)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"data": "Hello Golang gin-gonic & mongoDB",
		})
	})
	router.Run("localhost:6000")
}