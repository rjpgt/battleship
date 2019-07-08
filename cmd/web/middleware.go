package main

import (
	"fmt"
	"net/http"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "sameorigin")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) gameExists(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gameID := r.URL.Query().Get(":gameid")
		_, ok := app.gameModel.Games[gameID]
		if !ok {
			app.session.Put(r, "flash", "No such game or game has expired. Create a new game.")
			http.Redirect(w, r, "/start", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) canJoin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pgame, _ := app.gameModel.Games[r.URL.Query().Get(":gameid")]
		if len(pgame.Players) == 2 {
			app.session.Put(r, "flash", "Game is full. Start another.")
			http.Redirect(w, r, "/start", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) belongsToGame(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gameID := r.URL.Query().Get(":gameid")
		if gameID != app.session.GetString(r, "gameID") {
			app.session.Put(r, "flash", "No such game or you are not a part of the game. Start another.")
			http.Redirect(w, r, "/start", http.StatusSeeOther)
			return
		}

		pgame, _ := app.gameModel.Games[gameID]
		playerID := app.session.GetString(r, "playerID")
		_, ok := pgame.Players[playerID]
		if !ok {
			app.session.Put(r, "flash", "You are not a part of this game. Create a new game.")
			http.Redirect(w, r, "/start", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
