package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

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

	// db, err := sql.Open("mysql", "root:root1234@tcp(10.62.170.172:3306)/alert_db")
	db, err := sql.Open("mysql", "root:babihutan123@tcp(localhost:3306)/alert_db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	r.GET("/", func(c *gin.Context) {
		readAll, err := db.Query("SELECT * FROM alert_db.report")
		if err != nil {
			panic(err.Error())
		}

		defer readAll.Close()

		var result []map[string]interface{}

		for readAll.Next() {
			var (
				Name     string
				Link     string
				Severity string
				Date     string
				Message  string
				Host     string
				Owner    string
			)
			err := readAll.Scan(&Name, &Link, &Severity, &Date, &Message, &Host, &Owner)
			if err != nil {
				panic(err.Error())
			}

			result = append(result, map[string]interface{}{
				"Name":     Name,
				"Link":     Link,
				"Severity": Severity,
				"Date":     Date,
				"Message":  Message,
				"Host":     Host,
				"Owner":    Owner,
			})
		}

		c.JSON(http.StatusOK, gin.H{"result": result})
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

		t := time.Now()

		FullDate := strconv.Itoa(t.Year()) + " " + webhook.Result.DateTime

		insert, err := db.Query("INSERT INTO alert_db.report (Name, Link, Severity, Date, Message, Host, Owner) VALUES (?, ?, ?, ?, ?, ?, ?)", webhook.SearchName, webhook.ResultLink, webhook.Result.Severity, FullDate, webhook.Result.Raw, webhook.Result.Host, webhook.Owner)
		if err != nil {
			panic(err.Error())
		}

		defer insert.Close()

		// Respond to Splunk with a success message
		c.JSON(http.StatusOK, gin.H{"message": "Successfully processed Splunk webhook"})
	})

	r.Run(":8080")

}
