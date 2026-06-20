package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var password string
	if len(os.Args) > 1 {
		password = os.Args[1]
	} else {
		fmt.Print("Enter password: ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to read password:", err)
			os.Exit(1)
		}
		password = strings.TrimRight(line, "\r\n")
	}

	if password == "" {
		fmt.Fprintln(os.Stderr, "password is empty")
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hash failed:", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("password_hash:")
	fmt.Println(string(hash))
}
