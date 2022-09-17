package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (d *Ingest) registerBasicChecks() {

	// Sanity GET request
	d.Engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// / so things know we exist
	d.Engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ingestd",
		})
	})

	// metrics
	d.Engine.GET("/metrics", func(c *gin.Context) {
		gin.WrapH(promhttp.Handler())
	})
}

func (d *Ingest) registerNoRouteCheck() {
	d.Engine.NoRoute(func(c *gin.Context) {
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
		schema, table := d.parsePath(c.Request.URL.Path)
		if schema == "" || table == "" {
			log.Println("No route found")
			c.AbortWithStatusJSON(404, gin.H{
				"message": "No route found",
			})
			return
		}

		if d.canInsert(data) {
			switch d.getDBType() {
			case "mysql", "postgres":
				err = d.Insert(schema, table, data)
				if err != nil {
					d.abort(c, err)
					return
				}
			case "redis":
				err = d.Publish(d.makeChannelName(schema, table), data)
				if err != nil {
					d.abort(c, err)
					return
				}
			}
			c.JSON(200, gin.H{
				"message": "Success",
			})
		} else {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "Not Authorized",
			})
		}
	})
}

func (d *Ingest) abort(c *gin.Context, err error) {
	log.Println("Error inserting data to", d.getDBType())
	log.Println(err)
	c.AbortWithStatusJSON(500, gin.H{
		"message": fmt.Sprint(err),
	})
}
