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
	Key         string
	Login       string
	Password    string
	Description string
}

type Database struct {
	FileName string
	Password string
	Entries  map[string]Password
}

func New(fileName, password string) *Database {
	database = &Database{
		FileName: fileName,
		Password: password,
		Entries:  make(map[string]Password),
	}
	return database
}

func (d *Database) Load() (err error) {

	f, err := os.Open(d.FileName)
	defer f.Close()

	if os.IsNotExist(err) {
		return err
	}

	if err != nil {
		return fmt.Errorf("couldn't open db file %s: %s", d.FileName, err)
	}

	// FIXME that's weird solution

	var tries int = 0
	promptFunction := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if tries > 0 {
			return nil, fmt.Errorf("invalid password")
		}

		tries++
		return []byte(d.Password), nil
	}

	md, err := openpgp.ReadMessage(f, nil, promptFunction, nil)

	if err != nil {
		return fmt.Errorf("decryption failed: %s ", err)
	}

	bytes, err := ioutil.ReadAll(md.UnverifiedBody)

	if err != nil {
		return fmt.Errorf("reading decrypted message: %s", err)
	}

	fmt.Printf("Loading %s\n", d.FileName)
	err = json.Unmarshal(bytes, &d.Entries)
	if err != nil {
		return fmt.Errorf("unmarshalling failed: %s", err)
	}

	fmt.Printf("Got %d passwords\n", len(d.Entries))
	return nil
}

func (d *Database) Save() error {

	blob, err := json.MarshalIndent(&d.Entries, "", "    ")
	if err != nil {
		return fmt.Errorf("marshalling failed: %s", err)
	}

	f, err := os.Create(d.FileName)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("creating '%s' failed: %s", d.FileName, err)
	}

	writer, err := openpgp.SymmetricallyEncrypt(f, []byte(d.Password), nil, nil)

	if err != nil {
		return fmt.Errorf("encryption failed: %s", err)
	}

	_, err = writer.Write(blob)

	if err != nil {
		return fmt.Errorf("writing %s failed: %s", d.FileName, err)
	}

	writer.Close()

	fmt.Printf("Saved %s\n", d.FileName)
	return nil
}

func (d *Database) Add(key, login, pass, description string) {

	var p Password
	p.Key = key
	p.Login = login
	p.Password = pass
	p.Description = description

	d.Entries[key] = p
}

func (d *Database) Delete(key string) {
	delete(d.Entries, key)
}

func (d *Database) Get(key string) *Password {
	p, exists := d.Entries[key]
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

func (d *Database) Search(q string) []Password {
	var res = []Password{}

	for _, pass := range d.Entries {
		if searchMatch(pass, q) {
			res = append(res, pass)
		}
	}

	return res
}
