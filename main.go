package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"math/rand"
	"time"

	"github.com/howeyc/gopass"
	readline "gopkg.in/readline.v1"
)

type cmdResult int

const (
	ok cmdResult = iota
	noSuchCommand
	invalidNumberOfParameter
	missingKeyParameter
	nothingToShow
	alreadyExists
	quit
	abort
	otherError
)

var prompt = "npass> "
var commands map[string]func([]string) cmdResult
var bio *bufio.Reader

var recentSearchResults []Password
var recentKey *string
var database *Database

func init() {
	commands = make(map[string]func([]string) cmdResult)

	commands["help"] = helpCmd
	commands["add"] = addCmd
	commands["edit"] = editCmd
	commands["delete"] = delCmd
	commands["rename"] = renameCmd
	commands["list"] = searchCmd
	commands["ls"] = searchCmd
	commands["show"] = printCmd
	commands["cat"] = printCmd
	commands["quit"] = quitCmd
	commands["copy"] = copyCmd
	commands["generate"] = generateCmd
	commands["gen"] = generateCmd

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
	pass, err := gopass.GetPasswd()
	if err != nil {
		return nil
	}
	str := string(pass)
	return &str
}

func quitCmd(_ []string) cmdResult {
	return quit
}

func helpCmd(_ []string) cmdResult {
	fmt.Println("usage: npass")

	for cmd := range commands {
		fmt.Printf("%s\n", cmd)
	}
	return ok
}

func listCmd(params []string) cmdResult {
	var q = []string{""}
	return searchCmd(q)

}

func searchCmd(params []string) cmdResult {

	q := ""

	if len(params) > 0 {
		q = params[0]
	}

	recentSearchResults = database.Search(q)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for idx, pass := range recentSearchResults {
		fmt.Fprintf(w, "%d)\t%s \t-\t%s\t\n", idx, pass.Key, pass.Description)
	}

	w.Flush()

	return ok
}

func resetRecentResult() {
	recentSearchResults = []Password{}
	recentKey = nil
}

func toKey(s string) string {

	if i, err := strconv.Atoi(s); err == nil {
		if i >= 0 && i < len(recentSearchResults) {
			return recentSearchResults[i].Key
		}
	}

	return s
}

func addCmd(params []string) cmdResult {
	if len(params) != 1 {
		fmt.Println("Missing login parameter")
		return ok
	}

	key := params[0]

	if database.Get(key) != nil {
		return alreadyExists
	}

	recentKey = &key

	fmt.Print("Login: ")
	var login = line()

	if login == nil || *login == "" {
		return abort
	}

	//fmt.Print("Password: ")
	var password = readPassword("Password: ")

	if password == nil || *password == "" {
		return abort
	}

	fmt.Print("Description: ")
	var desc = line()

	if desc == nil {
		return abort
	}

	database.Add(key, *login, *password, *desc)
	err := database.Save()
	if err != nil {
		fmt.Printf("Error while saving %s\n", err)
	}
	resetRecentResult()
	return ok
}

func editCmd(params []string) cmdResult {

	p := findEntry(params)

	if p == nil {
		return missingKeyParameter
	}

	fmt.Printf("Login[%s]: ", p.Login)
	var login = line()

	if login == nil || *login == "" {
		login = &p.Login
	}

	//fmt.Print("Password: ")
	var password = readPassword("Password: ")

	if password == nil || *password == "" {
		password = &p.Password
	}

	fmt.Printf("Description[%s]: ", p.Description)
	var desc = line()

	if desc == nil || *desc == "" {
		desc = &p.Description
	}

	database.Add(p.Key, *login, *password, *desc)
	err := database.Save()
	if err != nil {
		fmt.Printf("Error while saving %s\n", err)
	}
	resetRecentResult()
	return ok
}

func delCmd(params []string) cmdResult {
	if len(params) != 1 {
		return missingKeyParameter
	}

	key := toKey(params[0])

	database.Delete(key)
	err := database.Save()
	if err != nil {
		fmt.Printf("Error while saving %s\n", err)
	}
	resetRecentResult()
	return ok
}

func renameCmd(params []string) cmdResult {
	if len(params) != 2 {
		return invalidNumberOfParameter
	}

	from := toKey(params[0])
	to := params[1]

	p := database.Get(from)

	if p == nil {
		return nothingToShow
	}

	database.Delete(p.Key)
	p.Key = to

	database.Add(p.Key, p.Login, p.Password, p.Description)
	err := database.Save()
	if err != nil {
		fmt.Printf("Error while saving %s\n", err)
	}
	return ok
}

func findEntry(params []string) *Password {
	var k string
	if len(params) != 1 {
		if recentKey == nil {
			k = toKey("0")
		} else {
			k = toKey(*recentKey)
		}
	} else {
		k = toKey(params[0])
	}

	res := database.Get(k)

	if res != nil {
		recentKey = &k
	}

	return res
}

func printCmd(params []string) cmdResult {

	p := findEntry(params)

	if p == nil {
		return nothingToShow
	}

	p.Print()

	return ok
}

func copyCmd(params []string) cmdResult {

	p := findEntry(params)

	if p == nil {
		return nothingToShow
	}

	err := copyToCliboard(p.Password)
	if err != nil {
		return otherError
	}
	fmt.Printf("Copied to clipboard.\n")

	return ok
}

func init() {
	rand.Seed(time.Now().Unix())
}

func shuffle(ary []byte) {

	for i := 0; i < 5; i++ {
		a, b := rand.Intn(len(ary)), rand.Intn(len(ary))

		tmp := ary[a]
		ary[a] = ary[b]
		ary[b] = tmp

	}

}

func generatePassword() string {
	len := 12
	res := make([]byte, len)

	for i := 0; i < len; i++ {

		var x int
		switch i % 10 {
		case 0:
			x = 48 + rand.Intn(10)
		case 1:
			x = 42 + rand.Intn(5)
		case 2:
			x = 65 + rand.Intn(26)
		default:
			x = 97 + rand.Intn(26)

		}

		res[i] = byte(x)
	}
	shuffle(res)
	return string(res)

}

func generateCmd(params []string) cmdResult {

	for i := 0; i < 10; i++ {
		fmt.Println(generatePassword())
	}

	return ok
}

func call(cmd string, params []string) cmdResult {

	p := findEntry([]string{cmd})
	if p != nil {
		p.Print()
		return ok
	}

	fn := commands[cmd]
	if fn == nil {
		return noSuchCommand
	}
	return fn(params)

}

func repl() {

	var items []readline.PrefixCompleterInterface
	for k := range commands {
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
				case ok:
					//fmt.Println("ok")
				case abort:
					fmt.Println("Aborted")
				case quit:
					return
				case missingKeyParameter:
					fmt.Println("Missing required key parameter")
				case invalidNumberOfParameter:
					fmt.Println("Invalid number of parameters")
				case noSuchCommand:
					fmt.Println("Unknown command:", cmd)
				case nothingToShow:
					fmt.Println("Nothing to display")
				case alreadyExists:
					fmt.Println("Already exists")
				default:
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

	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dbFileName := dirname + "/.npass.db"
	dbPassword := ""

	if !exists(dbFileName) {
		fmt.Printf("File '%s' doesn't not exists. Creating new database. Press control-c to abort.\n", dbFileName)
		p1 := readPassword("Password: ")

		if p1 == nil || *p1 == "" {
			fmt.Println("Aborted.")
			return
		}

		p2 := readPassword("Confirm password: ")

		if p2 == nil {
			fmt.Println("Aborted.")
			return
		}

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
		if p == nil {
			fmt.Println("Aborted.")
			return
		}
		dbPassword = *p
		database = New(dbFileName, dbPassword)
		err := database.Load()
		if err != nil {
			fmt.Printf("Unable to load the database '%s'\n", err)
			return
		}
	}

	repl()

}
