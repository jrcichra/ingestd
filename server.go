package main

import (
	"log"
	"strings"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// What makes up an ingestd server
type Ingest struct {
	*gin.Engine
	db  *sql.DB
	rdb *redis.Client
}

func (d *Ingest) Run() {
	d.Engine = gin.Default()
	dbtype := d.getDBType()
	if dbtype == "redis" {
		var err error
		d.rdb, err = d.connectToRedis(d.getDSN())
		if err != nil {
			log.Println("Error connecting to redis")
			log.Println(err)
			return
		}
	} else {
		d.db = d.reconnectToDB(d.getDSN(), dbtype)
	}
	// Register basic health checks
	d.registerBasicChecks()
	d.registerNoRouteCheck()
	//NoRoute replaces 404 with a route (expecting schema/table_name) with a JSON payload of what is to be inserted

	d.Engine.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func (d *Ingest) canInsert(data map[string]interface{}) bool {
	authenticated := d.isAuthorized(data)
	inAuthMode := d.inAuthMode()
	return (authenticated && inAuthMode) || (!authenticated && !inAuthMode) || (authenticated && !inAuthMode)
}

func (d *Ingest) parsePath(path string) (string, string) {
	//remove any leading slashes
	path = strings.TrimLeft(path, "/")
	//get the first two parts of the path
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
