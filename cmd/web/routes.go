package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

func (app *application) router() http.Handler {
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	dynamicMiddleware := alice.New(app.session.Enable)

	mux := pat.New()
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	mux.Get("/sse", dynamicMiddleware.ThenFunc(app.handleSse))
	mux.Get("/start", dynamicMiddleware.ThenFunc(app.startGameForm))
	mux.Post("/start", dynamicMiddleware.ThenFunc(app.startGame))
	mux.Get("/join/:gameid", dynamicMiddleware.Append(app.gameExists, app.canJoin).ThenFunc(app.joinGameForm))
	mux.Post("/join/:gameid", dynamicMiddleware.Append(app.gameExists, app.canJoin).ThenFunc(app.joinGame))
	mux.Get("/:gameid", dynamicMiddleware.Append(app.gameExists, app.belongsToGame).ThenFunc(app.playGameForm))
	mux.Post("/:gameid", dynamicMiddleware.Append(app.gameExists, app.belongsToGame).ThenFunc(app.playGame))

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	return standardMiddleware.Then(mux)
}
