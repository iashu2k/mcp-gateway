package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run ./cmd/passwordhash <password>")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(os.Args[1]),
		bcrypt.DefaultCost,
	)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	fmt.Println(string(hash))
}
