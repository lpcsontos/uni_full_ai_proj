/*
package main

import (
    "database/sql"
    "net/http"
    "os"
    "time"
	 "log"

	 "github.com/joho/godotenv"

    "github.com/go-chi/chi/v5"
    _ "github.com/go-sql-driver/mysql"

    "github.com/yourorg/recipemgr/internal/api"
    "github.com/yourorg/recipemgr/internal/repository"
    "github.com/yourorg/recipemgr/internal/usecase"
    "github.com/yourorg/recipemgr/pkg/jwt"
    "github.com/yourorg/recipemgr/pkg/logger"
)

func main() {
	 // load .env if present; ignore error in production
    if err := godotenv.Load(); err != nil {
        log.Println("no .env file found, relying on environment")
    }

    l := logger.New()
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        l.Fatal("DATABASE_URL not set")
    }
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        l.Fatalf("db open: %v", err)
    }
    db.SetConnMaxLifetime(time.Minute * 5)
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)

    // repositories
    repos := repository.NewMySQLRepository(db, l)

    // jwt manager
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        l.Fatal("JWT_SECRET not set")
    }
    jm := jwt.NewManager(secret, time.Minute*15, time.Hour*24*7) // access 15m, refresh 7d

    // usecases
    ucs := usecase.NewUsecases(repos, jm, l)

    // handlers
    h := api.NewHandler(ucs, l)

    r := chi.NewRouter()
    r.Use(logger.ChiMiddleware(l))
    r.Route("/api", func(r chi.Router) {
        h.RegisterRoutes(r)
    })

    addr := "localhost:8080"
    l.Infof("starting server on %s", addr)
    if err := http.ListenAndServe(addr, r); err != nil {
        l.Fatalf("server failed: %v", err)
    }
}

*/

package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/joho/godotenv"
    "github.com/rs/cors"

    "github.com/go-chi/chi/v5"
    _ "github.com/go-sql-driver/mysql"

    "github.com/yourorg/recipemgr/internal/api"
    "github.com/yourorg/recipemgr/internal/repository"
    "github.com/yourorg/recipemgr/internal/usecase"
    "github.com/yourorg/recipemgr/pkg/jwt"
    "github.com/yourorg/recipemgr/pkg/logger"
)

func main() {
    // load .env if present; ignore error in production
    if err := godotenv.Load(); err != nil {
        log.Println("no .env file found, relying on environment")
    }

    l := logger.New()

    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        l.Fatal("DATABASE_URL not set")
    }
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        l.Fatalf("db open: %v", err)
    }
    db.SetConnMaxLifetime(time.Minute * 5)
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)

    // repositories
    repos := repository.NewMySQLRepository(db, l)

    // jwt manager
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        l.Fatal("JWT_SECRET not set")
    }
    jm := jwt.NewManager(secret, time.Minute*15, time.Hour*24*7) // access 15m, refresh 7d

    // usecases
    ucs := usecase.NewUsecases(repos, jm, l)

    // handlers
    h := api.NewHandler(ucs, l)

    r := chi.NewRouter()
    r.Use(logger.ChiMiddleware(l))
    r.Route("/api", func(r chi.Router) {
        h.RegisterRoutes(r)
    })

    // CORS configuration
    // Adjust AllowedOrigins to match your frontend origin(s) exactly.
    // For local dev you can include both localhost and 127.0.0.1 variants.
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://127.0.0.1:5173", "http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
        AllowCredentials: true,
        Debug:            false,
    })

    // Wrap router with CORS handler
    handler := c.Handler(r)

    // Host and port from env with sensible defaults
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    host := os.Getenv("HOST")
    if host == "" {
        host = "0.0.0.0"
    }
    addr := fmt.Sprintf("%s:%s", host, port)

    l.Infof("starting server on %s", addr)
    if err := http.ListenAndServe(addr, handler); err != nil {
        l.Fatalf("server failed: %v", err)
    }
}
