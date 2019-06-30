package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golangcollege/sessions"
	"github.com/rjpgt/battleship/pkg/models"
)

type application struct {
	errorLog      *log.Logger
	gameModel     *models.GameModel
	infoLog       *log.Logger
	session       *sessions.Session
	templateCache map[string]*template.Template
}

const GameTimeout = 5
const MaxGames = 5

func main() {
	infoFile, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer infoFile.Close()
	infoLog := log.New(infoFile, "INFO\t", log.Ldate|log.Ltime)

	errFile, err := os.OpenFile("err.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer errFile.Close()
	errorLog := log.New(errFile, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte("2NJJssnekSBl3n0k@cg;S<B2rtleLPyw"))
	session.Lifetime = 12 * time.Hour
	session.HttpOnly = false
	session.Persist = false

	// This is our DB
	games := map[string]*models.Game{}

	app := &application{
		errorLog:      errorLog,
		gameModel:     &models.GameModel{Games: games},
		infoLog:       infoLog,
		session:       session,
		templateCache: templateCache,
	}

	//addr := ":21837"
	addr := ":8000"
	srv := &http.Server{
		Addr:     addr,
		ErrorLog: errorLog,
		Handler:  app.router(),
	}

	infoLog.Printf("Starting server on %s", addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
