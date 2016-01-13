package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func copyToCliboard(s string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(s)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("unable copy to cliboard: %s", err)
	}
	return nil
}
