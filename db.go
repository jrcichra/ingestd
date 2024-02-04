package main

import (
	"database/sql"
	"log"
	"sort"
	"strconv"
	"time"
)

func (d *Ingest) connectToDB(dsn string, dbtype string) (*sql.DB, error) {
	// connect to the database
	log.Println("Connecting with dbtype", dbtype)
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

func (d *Ingest) reconnectToDB(dsn string, dbtype string) *sql.DB {
	var db *sql.DB
	for {
		var err error
		db, err = d.connectToDB(dsn, dbtype)
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

func (d *Ingest) getDSN() string {
	config, err := d.readEnvironmentFile()
	if err != nil {
		log.Println("Error reading environment file")
		log.Println(err)
		return ""
	}
	//check for DSN key
	if _, ok := config["DSN"]; ok {
		return config["DSN"]
	} else {
		panic("dsn was empty. Is your configuration file set properly?")
	}
}

func (d *Ingest) getDBType() string {
	config, err := d.readEnvironmentFile()
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

func (d *Ingest) Insert(schema string, table string, data map[string]interface{}) error {
	var params []interface{}
	insert := "insert into " + schema + "." + table + " ("

	// Sort by key alphabetically so the database can use a statement over and over
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		// prepare the column and value for insert
		insert += k + ","
		params = append(params, data[k])
	}
	// Remove the last comma
	insert = insert[:len(insert)-1]
	//values
	insert += ") VALUES ("
	//variable for each value
	i := 1
	for range keys {
		//jonathandbriggs
		//Added Switch Case for pg/mysql. pg wants %n mysql wants ?
		switch d.getDBType() {
		case "postgres":
			insert += "$" + strconv.Itoa(i) + ","
		case "mysql":
			insert += "?" + ","
		}
		i++
	}
	// Remove the last comma
	insert = insert[:len(insert)-1]
	//end paren
	insert += ")"
	_, err := d.db.Exec(insert, params...)
	return err
}
