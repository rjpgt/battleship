package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/rjpgt/battleship/pkg/models"
)

type application struct {
	errorLog      *log.Logger
	gameModel     *models.GameModel
	infoLog       *log.Logger
	sessionStore  *sessions.CookieStore
	templateCache map[string]*template.Template
}

// GameTimeout is the maximum no. of hours a game is stored
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

	sessionStore := sessions.NewCookieStore([]byte("2NJJssnekSBl3n0k@cg;S<B2rtleLPyw"))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 5,
		HttpOnly: false,
	}

	// This is our DB
	games := map[string]*models.Game{}

	app := &application{
		errorLog:      errorLog,
		gameModel:     &models.GameModel{Games: games},
		infoLog:       infoLog,
		sessionStore:  sessionStore,
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
