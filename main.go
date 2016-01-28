package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/howeyc/gopass"
	readline "gopkg.in/readline.v1"
)

type CommandResult int

const (
	OK CommandResult = iota
	NoSuchCommand
	InvalidNumberOfParameter
	MissingKeyParameter
	NothingToShow
	Quit
	Abort
	Error
)

var prompt = "npass> "
var commands map[string]func([]string) CommandResult
var bio *bufio.Reader

var recentSearchResults []Password

var database *Database

func init() {
	commands = make(map[string]func([]string) CommandResult)

	commands["help"] = helpCmd
	commands["search"] = searchCmd
	commands["add"] = addCmd
	commands["delete"] = delCmd
	commands["rename"] = renameCmd
	commands["list"] = listCmd
	commands["print"] = printCmd
	commands["quit"] = quitCmd
	commands["copy"] = copyCmd

	bio = bufio.NewReader(os.Stdin)
	recentSearchResults = make([]Password, 0)

}

// commands

// https://github.com/chzyer/readline

func line() *string {

	data, hasMoreInLine, err := bio.ReadLine()

	if err != nil {
		fmt.Println(err)

		return nil
	}

	if hasMoreInLine == true {
		return nil
	}

	var l = string(data)
	return &l
}

func readPassword(msg string) *string {
	fmt.Print(msg)
	pass := string(gopass.GetPasswd())
	return &pass
}

func quitCmd(_ []string) CommandResult {
	return Quit
}

func helpCmd(_ []string) CommandResult {
	fmt.Println("usage: npass")

	for cmd, _ := range commands {
		fmt.Printf("%s\n", cmd)
	}
	return OK
}

func listCmd(params []string) CommandResult {
	var q = []string{""}
	return searchCmd(q)

}

func searchCmd(params []string) CommandResult {

	q := ""

	if len(params) > 0 {
		q = params[0]
	}

	recentSearchResults = database.Search(q)

	for idx, pass := range recentSearchResults {
		fmt.Printf("%d) %s - %s\n", idx, pass.Key, pass.Description)
	}
	return OK
}

func resetRecentResult() {
	recentSearchResults = []Password{}
}

func toKey(s string) string {

	if i, err := strconv.Atoi(s); err == nil {
		if i >= 0 && i < len(recentSearchResults) {
			return recentSearchResults[i].Key
		}
	}
	return s
}

func addCmd(params []string) CommandResult {
	if len(params) != 1 {
		fmt.Println("Missing login parameter")
		return OK
	}

	key := params[0]

	fmt.Print("Login: ")
	var login = line()

	if login == nil || *login == "" {
		return Abort
	}

	//fmt.Print("Password: ")
	var password = readPassword("Password: ")

	if password == nil || *password == "" {
		return Abort
	}

	fmt.Print("Description: ")
	var desc = line()

	if desc == nil {
		return Abort
	}

	database.Add(key, *login, *password, *desc)
	err := database.Save()
	if err != nil {
		fmt.Printf("Error while saving %s\n", err)
	}
	resetRecentResult()
	return OK
}

func delCmd(params []string) CommandResult {
	if len(params) != 1 {
		return MissingKeyParameter
	}

	key := toKey(params[0])

	database.Delete(key)
	err := database.Save()
	if err != nil {
		fmt.Printf("Error while saving %s\n", err)
	}
	resetRecentResult()
	return OK
}

func renameCmd(params []string) CommandResult {
	if len(params) != 2 {
		return InvalidNumberOfParameter
	}

	from := toKey(params[0])
	to := params[1]

	p := database.Get(from)

	if p == nil {
		return NothingToShow
	}

	database.Delete(p.Key)
	p.Key = to

	database.Add(p.Key, p.Login, p.Password, p.Description)
	err := database.Save()
	if err != nil {
		fmt.Printf("Error while saving %s\n", err)
	}
	return OK
}

func findEntry(params []string) *Password {
	var k string
	if len(params) != 1 {
		k = toKey("0")
	} else {
		k = toKey(params[0])
	}

	return database.Get(k)
}

func printCmd(params []string) CommandResult {

	p := findEntry(params)

	if p == nil {
		return NothingToShow
	}

	fmt.Printf("Key:         %s\n", p.Key)
	fmt.Printf("Login:       %s\n", p.Login)
	fmt.Printf("Password:    %s\n", p.Password)
	fmt.Printf("Description: %s\n", p.Description)
	return OK
}

func copyCmd(params []string) CommandResult {

	p := findEntry(params)

	if p == nil {
		return NothingToShow
	}

	err := copyToCliboard(p.Password)
	if err != nil {
		return Error
	}
	fmt.Printf("Copied to clipboard.\n")

	return OK
}

func call(cmd string, params []string) CommandResult {

	fn := commands[cmd]
	if fn == nil {
		return NoSuchCommand
	} else {
		return fn(params)
	}
}

func repl() {

	var items []*readline.PrefixCompleter
	for k, _ := range commands {
		items = append(items, readline.PcItem(k))
	}

	var completer = readline.NewPrefixCompleter(items...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       prompt,
		AutoComplete: completer,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		ary := strings.Split(string(line), " ")
		if len(ary) > 0 {
			cmd := ary[0]
			if cmd != "" {
				params := ary[1:]

				switch call(cmd, params) {
				case OK:
					//fmt.Println("OK")
				case Abort:
					fmt.Println("Aborted")
				case Quit:
					return
				case MissingKeyParameter:
					fmt.Println("Missing required key parameter")
				case InvalidNumberOfParameter:
					fmt.Println("Invalid number of parameters")
				case NoSuchCommand:
					fmt.Println("Unknown command:", cmd)
				case NothingToShow:
					fmt.Println("Nothing to display")
				case Error:
					fmt.Println("Error")
				}
			}
		}
	}

}

func usage() {
	fmt.Println(`usage:

npass passwords.db`)
}

func exists(f string) bool {
	if _, err := os.Stat(f); os.IsNotExist(err) {
		return false
	}
	return true
}
func main() {

	var err error

	if len(os.Args) != 2 {
		usage()
		return
	}

	dbFileName := os.Args[1]
	dbPassword := ""

	if !exists(dbFileName) {
		fmt.Printf("File '%s' doesn't not exists. Creating new database. Press control-c to abort.\n", dbFileName)
		p1 := readPassword("Password: ")
		p2 := readPassword("Confirm password: ")

		if *p1 != *p2 {
			fmt.Println("Passwords don't match. Abort.")
			return
		}
		dbPassword = *p1

		database = New(dbFileName, dbPassword)
		err := database.Save()
		if err != nil {
			fmt.Printf("Error while saving %s \n", err)
			return
		}
	} else {

		fmt.Println("Please enter a password")
		p := readPassword("Password: ")
		dbPassword = *p
		database = New(dbFileName, dbPassword)
		err := database.Load()
		if err != nil {
			fmt.Printf("Unable to load the database '%s'\n", err)
			return
		}
	}

	repl()

	err = database.Save()
	if err != nil {
		fmt.Printf("Error while saving %s \n", err)
	}
}
