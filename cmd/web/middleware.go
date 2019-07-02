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
			session, _ := app.sessionStore.Get(r, "btlship-session")
			session.AddFlash("No such game or game has expired")
			session.Save(r, w)
			//http.Redirect(w, r, "/btlship/start", http.StatusSeeOther)
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
			session, _ := app.sessionStore.Get(r, "btlship-session")
			session.AddFlash("Game is full. Start another.")
			session.Save(r, w)
			//http.Redirect(w, r, "/btlship/start", http.StatusSeeOther)
			http.Redirect(w, r, "/start", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) belongsToGame(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gameID := r.URL.Query().Get(":gameid")
		session, _ := app.sessionStore.Get(r, "btlship-session")
		sessionGameID, _ := session.Values["gameID"].(string)
		if gameID != sessionGameID {
			session.AddFlash("No such game or you are not a part of the game. Start another.")
			session.Save(r, w)
			//http.Redirect(w, r, "/btlship/start", http.StatusSeeOther)
			http.Redirect(w, r, "/start", http.StatusSeeOther)
			return
		}

		pgame, _ := app.gameModel.Games[gameID]
		playerID, _ := session.Values["playerID"].(string)
		_, ok := pgame.Players[playerID]
		if !ok {
			session.AddFlash("You are not a part of this game. Create a new game.")
			//http.Redirect(w, r, "/btlship/start", http.StatusSeeOther)
			http.Redirect(w, r, "/start", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
