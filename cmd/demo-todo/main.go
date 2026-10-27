package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/adesokanayo/demo-todo/docs"
	"github.com/adesokanayo/demo-todo/internal/todo"
)

// @title           Demo TODO API
// @version         1.0
// @description     Create, read, and update todos stored in Postgres.
// @BasePath        /

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	store := todo.NewPostgres(db)
	if err := store.Ping(ctx); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Print("postgres connected and todos table ready")

	mux := http.NewServeMux()
	todo.Register(mux, store)
	mux.Handle("GET /swagger", http.RedirectHandler("/swagger/index.html", http.StatusFound))
	mux.Handle("GET /swagger/", docs.Handler())

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
