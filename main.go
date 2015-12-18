package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var prompt = "npass> "
var commands map[string]func([]string)
var bio *bufio.Reader

var recentSearchResults []Password

func init() {
	commands = make(map[string]func([]string))

	commands["help"] = helpCmd
	commands["search"] = searchCmd
	commands["add"] = addCmd
	commands["delete"] = delCmd
	commands["rename"] = renameCmd
	commands["list"] = listCmd
	commands["print"] = printCmd

	bio = bufio.NewReader(os.Stdin)
	recentSearchResults = make([]Password, 0)
}

// commands

// https://github.com/chzyer/readline

func readline() *string {

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

func helpCmd(_ []string) {
	fmt.Println("usage: npass")

	for cmd, _ := range commands {
		fmt.Printf("%s\n", cmd)
	}

}

func listCmd(params []string) {
	var q = []string{""}
	searchCmd(q)
}

func searchCmd(params []string) {

	q := ""

	if len(params) > 0 {
		q = params[0]
	}

	recentSearchResults = search(q)

	for idx, pass := range recentSearchResults {
		fmt.Printf("%d) %s - %s\n", idx, pass.Login, pass.Description)
	}
}

func resetRecentResult() {
	recentSearchResults = []Password{}
}

func toLogin(s string) string {

	if i, err := strconv.Atoi(s); err == nil {
		if i >= 0 && i < len(recentSearchResults) {
			return recentSearchResults[i].Login
		}
	}
	return s
}

func addCmd(params []string) {
	if len(params) != 1 {
		fmt.Println("Missing login parameter")
		return
	}

	login := params[0]

	fmt.Print("Password: ")
	var password = readline()

	if password == nil || *password == "" {
		fmt.Println("Aborted.")
		return
	}

	fmt.Print("Description: ")
	var desc = readline()

	if desc == nil {
		fmt.Println("Aborted.")
		return
	}

	add(login, *password, *desc)
	save()
	resetRecentResult()
}

func delCmd(params []string) {
	if len(params) != 1 {
		fmt.Println("Missing key parameter")
		return
	}

	login := toLogin(params[0])

	del(login)
	save()
	resetRecentResult()
}

func renameCmd(params []string) {
	if len(params) != 2 {
		fmt.Println("Wrong number of parameters")
	}

	from := toLogin(params[0])
	to := params[1]

	p := get(from)

	del(p.Login)
	p.Login = to

	add(p.Login, p.Password, p.Description)
	save()
}

func printCmd(params []string) {
	if len(params) != 1 {
		fmt.Println("Wrong number of parameters")
	}

	k := toLogin(params[0])

	p := get(k)

	fmt.Printf("Login:       %s\n", p.Login)
	fmt.Printf("Password:    %s\n", p.Password)
	fmt.Printf("Description: %s\n", p.Description)
}

func call(cmd string, params []string) {

	fn := commands[cmd]
	if fn == nil {
		fmt.Printf("No such command '%s' \n", cmd)
	} else {
		fn(params)
	}
}

func repl() {

	for {

		fmt.Print(prompt)

		line, hasMoreInLine, err := bio.ReadLine()

		if err != nil {
			fmt.Println(err)
			break
		}

		if hasMoreInLine == true {
			break
		}

		ary := strings.Split(string(line), " ")
		if len(ary) > 0 {
			cmd := ary[0]
			params := ary[1:]
			call(cmd, params)
		}
	}
}

func main() {

	load()
	repl()
	save()

}
