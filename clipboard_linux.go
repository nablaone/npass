package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func copyToClipboard(s string) err {
	cmd := exec.Command("cat")
	cmd.Stdin = strings.NewReader(s)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("unable copy to cliboard: %s", err)
	}
	return nil
}
