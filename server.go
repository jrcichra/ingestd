package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func connectToDB(dsn string, dbtype string) (*sql.DB, error) {
	// connect to the database
	log.Println("Connecting with dbtype", dbtype, "and dsn", dsn)
	db, err := sql.Open(dbtype, dsn)
	if err == nil {
		log.Println("Connected to database. Doing a Ping()")
		err = db.Ping()
		if err != nil {
			log.Println(err)
		}
		log.Println("Did a Ping()")
	} else {
		log.Println("Error connecting to database after Open", err)
	}
	return db, err
}

func reconnectToDB(dsn string, dbtype string) *sql.DB {
	var db *sql.DB
	for {
		var err error
		db, err = connectToDB(dsn, dbtype)
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

func connectToRedis(dsn string) (*redis.Client, error) {
	// parse the dsn for redis
	split := strings.Split(dsn, ":")
	host := split[0]
	password := split[1]
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       0,
	})
	return rdb, nil
}

func inAuthorizedKeys(key string) bool {
	//check if AUTHORIZED_KEYS is in the environment file
	//get the homedir of the user for the default
	homedir := os.Getenv("HOME")
	path := homedir + "/.ssh/authorized_keys"
	config, err := readEnvironmentFile()
	if err != nil {
		log.Println(err)
		return false
	}
	if _, ok := config["AUTHORIZED_KEYS"]; ok {
		//check if the key is in AUTHORIZED_KEYS
		path = config["AUTHORIZED_KEYS"]
	}
	authorizedKeys, err := os.Open(path)
	//check if key is in authorized keys
	if err != nil {
		log.Println("Error opening authorized keys file")
		log.Println(err)
		return false
	}
	defer authorizedKeys.Close()
	// scan each line of the authorized keys file
	scanner := bufio.NewScanner(authorizedKeys)
	for scanner.Scan() {
		// check if the key is in the line
		if strings.Contains(scanner.Text(), key) {
			return true
		}
	}
	// if the key is not in the authorized keys file, return an error
	log.Println("Key not found in authorized keys")
	return false
}

func getDSN() string {
	config, err := readEnvironmentFile()
	if err != nil {
		log.Println("Error reading environment file")
		log.Println(err)
		return ""
	}
	//check for DSN key
	if _, ok := config["DSN"]; ok {
		log.Println(config["DSN"])
		return config["DSN"]
	} else {
		panic("dsn was empty. Is your configuration file set properly?")
	}
}

func readEnvironmentFile() (map[string]string, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	// log.Println("path=", path)
	f, err := os.Open(path + "/config.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m := make(map[string]string)
	s := bufio.NewScanner(f)
	//read each line and separate the key and value which is separated by a =
	// ignore lines that start with a #
	// ignore lines that are empty
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "\n") {
			continue
		}
		if strings.HasPrefix(line, "\r") {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid environment file format")
		}
		m[kv[0]] = kv[1]
	}
	return m, nil
}

func inAuthMode() bool {
	config, err := readEnvironmentFile()
	if err != nil {
		log.Println("Error reading environment file")
		log.Println(err)
		return false
	}
	//check for AUTH_MODE key
	if val, ok := config["AUTH_MODE"]; ok {
		if strings.ToLower(val) == "true" || strings.ToLower(val) == "yes" {
			return true
		} else if strings.ToLower(val) == "false" || strings.ToLower(val) == "no" {
			return false
		} else {
			log.Println("auth_mode was not set to true or false. Defaulting to false")
			return false
		}
	}
	return false
}

func getDBType() string {
	config, err := readEnvironmentFile()
	if err != nil {
		log.Println("Error reading environment file")
		log.Println(err)
		return ""
	}
	//check for DB_TYPE key
	if val, ok := config["DB"]; ok {
		return val
	} else {
		panic("db was empty. Is your configuration file set properly?")
	}
}

func main() {

	var db *sql.DB
	var rdb *redis.Client

	dbtype := getDBType()
	if dbtype == "redis" {
		var err error
		rdb, err = connectToRedis(getDSN())
		if err != nil {
			log.Println("Error connecting to redis")
			log.Println(err)
			return
		}
	} else {
		db = reconnectToDB(getDSN(), dbtype)
	}

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

	//NoRoute replaces 404 with a route (expecting schema/table_name) with a JSON payload of what is to be inserted
	r.NoRoute(func(c *gin.Context) {
		// Handle internal errors by sending the error to the client
		defer func() {
			if err := recover(); err != nil {
				log.Println("Caught 500:", err)
				c.AbortWithStatusJSON(500, gin.H{
					"message": fmt.Sprint(err),
				})
			}
		}()

		// Get the json payload
		data := make(map[string]interface{})
		err := c.BindJSON(&data)
		if err != nil {
			log.Println("Error parsing JSON")
			log.Println(err)
			c.AbortWithStatusJSON(400, gin.H{
				"message": fmt.Sprint(err),
			})
			return
		}
		// Get the route
		route := c.Request.URL.String()
		// Get schema / table
		parts := strings.SplitN(route, "/", 3)
		schema := parts[1]
		table := parts[2]

		// Loop through keys in the interface
		ins := "insert into " + schema + "." + table + " ("

		// Put in the column names
		var params []interface{}
		authenticated := false
		for k, v := range data {
			//check if they passed in a public key
			if k == "_public_key" && inAuthMode() {
				// make sure the value is a string
				if v2, ok := v.(string); ok {
					if inAuthorizedKeys(v2) {
						// if the key is in authorized keys and AUTH_MODE is true, set authenticated to true
						authenticated = true
					} else {
						// if the key is not in authorized keys and we're in AUTH_MODE, return an error
						log.Println("Key not in authorized keys and AUTH_MODE is false")
						c.AbortWithStatusJSON(401, gin.H{
							"message": "unauthorized",
						})
					}
				}
			} else if k != "_public_key" {
				// prepare the column and value for insert
				ins += k + ","
				params = append(params, v)
			} else {
				// not sure how we got here. Skipping.
			}
		}
		// Remove the last comma
		ins = ins[:len(ins)-1]
		//values
		ins += ") VALUES ("
		//variable for each value
		i := 1
		for range data {
			//jonathandbriggs
			//Added Switch Case for pg/mysql. pg wants %n mysql wants ?
			switch getDBType() {
			case "postgres":
				ins += "$" + strconv.Itoa(i) + ","
			case "mysql":
				ins += "?" + ","
			}
			i++
		}
		// Remove the last comma
		ins = ins[:len(ins)-1]
		//end paren
		ins += ")"

		// before executing, make sure the authentication status matches the AUTH_MODE or we're authenticated but AUTH_MODE is false
		if (authenticated && inAuthMode()) || (!authenticated && !inAuthMode()) || (authenticated && !inAuthMode()) {
			//run
			fmt.Println("Insert statement:", ins)
			fmt.Println("params:")
			fmt.Println(params...)
			// Deprecated: drivers shoudl implement StmtExecContext instead (or additionally).
			// Exec(args []Value) (Result, error)
			_, err := db.Exec(ins, params...)
			if err != nil {
				fmt.Println(err)
				c.AbortWithStatusJSON(503, gin.H{
					"message": fmt.Sprint(err),
				})
			} else {
				c.Status(200)
			}
		} else {
			log.Println("Authentication failed")
			c.AbortWithStatusJSON(401, gin.H{
				"message": "unauthorized",
			})
		}

	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
