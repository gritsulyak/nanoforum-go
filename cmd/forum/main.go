package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gritsulyak/nanoforum-go/internal/config"
	"github.com/gritsulyak/nanoforum-go/internal/db"
	"github.com/gritsulyak/nanoforum-go/internal/handlers"
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
	posts := repository.NewPostRepo(conn)

	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	h := handlers.New(users, posts, tmpl)

	http.HandleFunc("/", h.Forum)
	http.HandleFunc("/login", h.Login)
	http.HandleFunc("/logout", h.Logout)

	log.Println("nanoforum running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
