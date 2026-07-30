package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gritsulyak/nanoforum-go/internal/auth"
	"github.com/gritsulyak/nanoforum-go/internal/config"
	"github.com/gritsulyak/nanoforum-go/internal/db"
	"github.com/gritsulyak/nanoforum-go/internal/repository"
)

func main() {
	conn, err := db.New(config.DBPath())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing DB connection: %v", err)
		}
	}()

	users := repository.NewUserRepo(conn)
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Login: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if username == "" || password == "" {
		log.Fatal("Login and password cannot be empty")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal(err)
	}

	if err := users.Create(username, hash); err != nil {
		log.Fatalf("Error creating user (possibly username is taken): %v", err)
	}

	fmt.Printf("User %s successfully created and saved to forum.db!\n", username)
}
