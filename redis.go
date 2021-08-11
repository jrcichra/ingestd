package main

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

func (d *Ingest) connectToRedis(dsn string) (*redis.Client, error) {
	// parse the dsn for redis
	split := strings.Split(dsn, ",")
	host := split[0]
	password := split[1]
	database := split[2]
	// convert the database to an int
	db, err := strconv.Atoi(database)
	if err != nil {
		log.Println("Error converting database to int")
		log.Println(err)
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})
	return rdb, nil
}

func (d *Ingest) Publish(channel string, data map[string]interface{}) error {
	// convert data to a string
	dataStr, err := json.Marshal(data)
	if err != nil {
		log.Println("Error converting data to string")
		log.Println(err)
		return err
	}
	// send the data to redis

	result := d.rdb.Publish(channel, string(dataStr))
	if result.Err() != nil {
		log.Println("Error publishing to channel")
		log.Println(result.Err())
		return result.Err()
	}
	return nil
}

func (d *Ingest) makeChannelName(schema string, table string) string {
	return schema + "." + table
}
