package main

import (
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
	"time"

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
	basePath := config.BasePath()
	h := handlers.New(users, posts, tmpl, basePath)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Forum)
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/logout", h.Logout)

	if config.PprofEnabled() {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		log.Println("pprof endpoints enabled under /debug/pprof/")
	}

	var rootHandler http.Handler = mux
	if basePath != "" {
		rootHandler = http.StripPrefix(basePath, mux)
	}

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           rootHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Forum run: http://localhost:8080%s", basePath)
	log.Fatal(srv.ListenAndServe())
}
