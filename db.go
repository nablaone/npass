package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"

	"golang.org/x/crypto/openpgp"
)

type Password struct {
	Login       string
	Password    string
	Description string
}

type Database map[string]Password

var database Database

var dbName = ".npass.db"
var dbFileName string
var dbPassword = "secret" // TODO ask user

func init() {
	u, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dbFileName = u.HomeDir + "/" + dbName
}

func password() string {
	return dbPassword
}

func create() {
	fmt.Println("creating new db")
	database = make(Database)
	save()
}

func load() (err error) {

	f, err := os.Open(dbFileName)
	defer f.Close()

	if os.IsNotExist(err) {
		create()
		return nil
	}

	promptFunction := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		return []byte(password()), nil
	}

	md, err := openpgp.ReadMessage(f, nil, promptFunction, nil)

	if err != nil {
		return fmt.Errorf("decryption failed: %s ", err)
	}

	bytes, err := ioutil.ReadAll(md.UnverifiedBody)

	if err != nil {
		return fmt.Errorf("reading decrypted message: %s", err)
	}

	fmt.Printf("Loading %s\n", dbFileName)
	err = json.Unmarshal(bytes, &database)
	if err != nil {
		return fmt.Errorf("unmarshalling failed: %s", err)
	}

	fmt.Printf("Got %d passwords\n", len(database))
	return nil
}

func save() error {

	blob, err := json.Marshal(&database)
	if err != nil {
		return fmt.Errorf("marshalling failed: %s", err)
	}

	f, err := os.Create(dbFileName)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("creating '%s' failed: %s", dbFileName, err)
	}

	writer, err := openpgp.SymmetricallyEncrypt(f, []byte(password()), nil, nil)

	if err != nil {
		return fmt.Errorf("encryption failed: %s", err)
	}

	_, err = writer.Write(blob)

	if err != nil {
		return fmt.Errorf("writing %s failed: %s", dbFileName, err)
	}

	writer.Close()

	fmt.Printf("Saved %s\n", dbFileName)
	return nil
}

func add(login, pass, description string) {

	var p Password
	p.Login = login
	p.Password = pass
	p.Description = description

	database[login] = p
}

func del(login string) {
	delete(database, login)
}

func get(login string) Password {
	return database[login]
}

func searchMatch(pass Password, q string) bool {

	if q == "" {
		return true
	}

	return strings.Contains(pass.Login, q) ||
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
