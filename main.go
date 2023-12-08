package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type splunkWebhook struct {
	Owner      string `json:"owner"`
	SearchName string `json:"search_name"`
	ResultLink string `json:"results_link"`
	Result     Result `json:"result"`
}

type Result struct {
	Severity  string `json:"aseverity"`
	Severity2 string `json:"Severity"`
	Raw       string `json:"_raw"`
	HostName  string `json:"hostname"`
	Host      string `json:"host"`
	DateTime  string `json:"date_time"`
}

func main() {
	// Connect to the database

	r := gin.Default()
	r.Use(cors.Default())

	// db, err := sql.Open("mysql", "root:root1234@tcp(10.62.170.172:3306)/alert")
	// if err != nil {
	// 	panic(err.Error())
	// }
	// defer db.Close()

	// err = db.Ping()
	// if err != nil {
	// 	panic(err.Error())
	// }

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

		// Parse the JSON payload into a splunkWebhook struct
		var webhook splunkWebhook
		err = json.Unmarshal(body, &webhook)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse JSON payload"})
			return
		}

		if webhook.Result.Severity == "" {
			if webhook.Result.Severity2 == "" {
				webhook.Result.Severity = "HIGH"
			} else {
				webhook.Result.Severity = webhook.Result.Severity2
			}
		}

		// Respond to Splunk with a success message
		c.JSON(http.StatusOK, gin.H{"owner": webhook.Owner, "search_name": webhook.SearchName, "results_link": webhook.ResultLink, "severity": webhook.Result.Severity, "hostname": webhook.Result.HostName, "host": webhook.Result.Host, "date_time": webhook.Result.DateTime, "message": webhook.Result.Raw})
	})

	r.Run(":8080")

}
