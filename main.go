package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World!"})
	})

	r.POST("/splunk-webhook", func(c *gin.Context) {
		// Read the JSON payload from the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		// Process the incoming JSON payload (you can parse it or do other actions)
		fmt.Println("Received Splunk webhook data:", string(body))

		// Respond to Splunk with a success message
		c.JSON(http.StatusOK, gin.H{"message": "Webhook data received successfully"})
	})

	r.Run(":8080")
}
