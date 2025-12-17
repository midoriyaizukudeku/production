package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"dream.website/internal/model"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       model.SnippetModel
	Users          model.UserModelInterface
	Templatecache  map[string]*template.Template
	FormDecoder    *form.Decoder
	SessionManager *scs.SessionManager
}

func main() {
	// Load environment variables from .env file
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Starting server on %s", addr)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime|log.Ldate)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := OpenDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	Templatecache, err := NewTemplatecache()
	if err != nil {
		errorLog.Fatal(err)
	}
	formDecoder := form.NewDecoder()

	SessionManager := scs.New()
	SessionManager.Store = mysqlstore.New(db)
	SessionManager.Lifetime = 12 * time.Hour
	SessionManager.Cookie.Secure = os.Getenv("RAILWAY_ENVIRONMENT") != ""

	app := &Application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		Users:          &model.UserModel{DB: db},
		snippets:       &model.SnipppetModel{DB: db},
		Templatecache:  Templatecache,
		FormDecoder:    formDecoder,
		SessionManager: SessionManager,
	}

	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     errorLog,
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
