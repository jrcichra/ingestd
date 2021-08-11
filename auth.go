package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func (d *Ingest) inAuthorizedKeys(key string) bool {
	//check if AUTHORIZED_KEYS is in the environment file
	//get the homedir of the user for the default
	homedir := os.Getenv("HOME")
	path := homedir + "/.ssh/authorized_keys"
	config, err := d.readEnvironmentFile()
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

func (d *Ingest) inAuthMode() bool {
	config, err := d.readEnvironmentFile()
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

func (d *Ingest) hasPublicKey(data map[string]interface{}) bool {
	_, ok := data["public_key"]
	return ok
}

func (d *Ingest) isAuthorized(data map[string]interface{}) bool {
	if d.hasPublicKey(data) && d.inAuthMode() {
		return d.inAuthorizedKeys(data["public_key"].(string))
	}
	return false
}
