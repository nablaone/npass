package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
)

type Password struct {
	Key         string
	Password    string
	Description string
}

type Database map[string]Password

var database Database

var dbName = ".npass.db"
var dbFileName string

func init() {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dbFileName = u.HomeDir + "/" + dbName
}

func create() {
	fmt.Println("creating new db")
	database = make(Database)
	save()
}

func load() (err error) {

	blob, err := ioutil.ReadFile(dbFileName)

	if os.IsNotExist(err) {
		create()
		return nil
	}

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Loading %s\n", dbFileName)
	err = json.Unmarshal(blob, &database)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Got %d passwords\n", len(database))
	return nil
}

func save() {

	blob, err := json.Marshal(&database)
	if err != nil {
		log.Fatal(err)
	}

	ioutil.WriteFile(dbFileName, blob, 0600)
	fmt.Printf("Saved %s\n", dbFileName)
}

func add(key, pass, description string) {

	var p Password
	p.Key = key
	p.Password = pass
	p.Description = description

	database[key] = p
}

func del(key string) {
	delete(database, key)
}

func get(key string) Password {
	return database[key]
}

func searchMatch(pass Password, q string) bool {

	if q == "" {
		return true
	}

	return strings.Contains(pass.Key, q) ||
		strings.Contains(pass.Description, q)
}

func search(q string) []Password {
	var res = []Password{}

	for _, pass := range database {
		if searchMatch(pass, q) {
			res = append(res, pass)
		}
	}

	return res
}
