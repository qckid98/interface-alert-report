package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
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

type Acknowledge struct {
	Id       int32  `json:"id"`
	Name     string `json:"Name"`
	Link     string `json:"Link"`
	Severity string `json:"Severity"`
	Date     string `json:"Date"`
	Message  string `json:"Message"`
	Host     string `json:"Host"`
	Owner    string `json:"Owner"`
	AckTime  string `json:"AckTime"`
}

func main() {
	// Connect to the database

	r := gin.Default()
	r.Use(cors.Default())

	db, err := sql.Open("mysql", "root:root1234@tcp(10.62.170.172:3306)/alert_db")
	// db, err := sql.Open("mysql", "root:babihutan123@tcp(localhost:3306)/alert_db")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	r.POST("/", func(c *gin.Context) {
		readAll, err := db.Query("SELECT * FROM alert_db.tester_db")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
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
				id       string
			)
			err := readAll.Scan(&Name, &Link, &Severity, &Date, &Message, &Host, &Owner, &id)
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
				"id":       id,
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

		insert, err := db.Query("INSERT INTO alert_db.tester_db (Name, Link, Severity, Date, Message, Host, Owner) VALUES (?, ?, ?, ?, ?, ?, ?)", webhook.SearchName, webhook.ResultLink, webhook.Result.Severity, webhook.Result.DateTime, webhook.Result.Raw, webhook.Result.Host, webhook.Owner)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer insert.Close()

		// Respond to Splunk with a success message
		c.JSON(http.StatusOK, gin.H{"message": "Successfully processed Splunk webhook"})
	})

	r.POST("/acknowledge/:id", func(c *gin.Context) {
		id := c.Param("id")

		var (
			Name     string
			Link     string
			Severity string
			Date     string
			Message  string
			Host     string
			Owner    string
			Id       string
		)

		ack := db.QueryRow("SELECT * FROM tester_db WHERE id = ?", id)

		err = ack.Scan(&Name, &Link, &Severity, &Date, &Message, &Host, &Owner, &Id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "No data found with id number " + id})
		}

		if err != nil {
			panic(err.Error())
		}

		datastring := Date

		layout := "2006-01-02 15:04:05"

		parsedTime, err := time.Parse(layout, datastring)
		if err != nil {
			panic(err.Error())
		}

		t := time.Now().UTC()

		difference := t.Sub(parsedTime)
		ackTime := difference.Minutes()

		insert, err := db.Query("INSERT INTO alert_db.report (Name, Link, Severity, Date, Message, Host, Owner, AckTime) VALUES (?,?,?,?,?,?,?,?)", Name, Link, Severity, Date, Message, Host, Owner, ackTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		}

		defer insert.Close()

		delete, err := db.Query("DELETE FROM tester_db WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Error when deleting"})
		}

		defer delete.Close()
		c.JSON(http.StatusOK, gin.H{"message": "Successfully Acknowledge an alert " + t.String() + " " + parsedTime.String()})
	})

	r.Run(":8080")

}
