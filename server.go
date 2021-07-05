package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func connectToDB(dsn *string) (*sql.DB, error) {
	// connect to the database
	// db, err := sql.Open("mysql", *dsn)
	db, err := sql.Open("postgres", *dsn)
	if err == nil {
		err = db.Ping()
	}
	return db, err
}

func reconnectToDB(dsn *string) *sql.DB {
	var db *sql.DB
	for {
		var err error
		db, err = connectToDB(dsn)
		if err != nil {
			log.Println(err)
			db.Close()
			time.Sleep(time.Duration(1) * time.Second)
			err = nil
		} else {
			break
		}

	}
	return db
}

func getDSN() *string {
	// Read the config file and return the DSN
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	log.Println("path=", path)
	f, err := os.Open(path + "/config.txt")
	if err != nil {
		panic(err)
	}
	fscanner := bufio.NewScanner(f)
	var dsn string
	for fscanner.Scan() {
		dsn = fscanner.Text()
		break
	}
	if dsn == "" {
		panic("dsn was empty. Is your configuration file set properly?")
	}

	log.Println("dsn=", dsn)
	return &dsn
}

func main() {

	db := reconnectToDB(getDSN())

	r := gin.Default()

	// Sanity GET request
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// / so things know we exist
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ingestd",
		})
	})

	//NoRoute replaces 404 with a route (expecting database/table_name) with a JSON payload of what is to be inserted
	r.NoRoute(func(c *gin.Context) {
		// Handle internal errors by sending the error to the client
		defer func() {
			if err := recover(); err != nil {
				log.Println("Caught 500:", err)
				c.AbortWithStatusJSON(500, err)
			}
		}()

		// Get the json payload
		data := make(map[string]interface{})
		c.BindJSON(&data)
		// Get the route
		route := c.Request.URL.String()
		// Get database / table
		parts := strings.SplitN(route, "/", 3)
		database := parts[1]
		table := parts[2]

		// Print debug
		// spew.Dump(data)
		// fmt.Println(route)

		// Loop through keys in the interface
		prepare := "insert " + database + "." + table + " set "
		// Add a question mark for each key and prepare the data in an array of interfaces
		var params []interface{}
		for k, v := range data {
			prepare += k + " = ?,"
			params = append(params, v)
		}
		// Remove the last comma
		prepare = prepare[:len(prepare)-1]

		// Prepare the statement
		// log.Println("prepare=", prepare)
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		defer tx.Rollback()
		stmt, err := tx.Prepare(prepare)
		if err != nil {
			panic(err)
		}

		_, err = stmt.Exec(params...)
		if err != nil {
			panic(err)
		}
		//commit
		err = tx.Commit()
		if err != nil {
			panic(err)
		}
		// Clear the prepared statement
		stmt.Close()

		// db.Exec("INSERT INTO ?.? VALUES ")
		c.Status(200)
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
