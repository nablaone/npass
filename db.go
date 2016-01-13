package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
)

type Password struct {
	Login       string
	Password    string
	Description string
}

// TODO refactor to a struct
type Database map[string]Password

var database Database

var dbFileName string
var dbPassword string = ""

func password() string {
	return dbPassword
}

func create() {
	database = make(Database)
	save()
}

func exists() bool {
	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		return false
	}
	return true
}

func load() (err error) {

	f, err := os.Open(dbFileName)
	defer f.Close()

	if os.IsNotExist(err) {
		return err
	}

	if err != nil {
		return fmt.Errorf("couldn't open db file %s: %s", dbFileName, err)
	}

	// FIXME that's weird solution

	var tries int = 0
	promptFunction := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if tries > 0 {
			return nil, fmt.Errorf("invalid password")
		}

		tries++
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

func get(login string) *Password {
	p, exists := database[login]
	if exists {
		return &p
	}
	return nil
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
