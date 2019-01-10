package io

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func GetUsernamePassword() (username string, password string) {
	reader := bufio.NewReader(os.Stdin)
	username = ""
	for len(username) == 0 {
		fmt.Print("Enter username: ")
		username, _ = reader.ReadString('\n')
		username = strings.TrimSpace(username)
	}

	password = ""
	for len(password) == 0 {
		fmt.Print("Enter password: ")
		bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
		password = string(bytePassword)
	}

	return username, password
}
