package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rjpgt/battleship/pkg/forms"
	"github.com/rjpgt/battleship/pkg/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	session, _ := app.sessionStore.Get(r, "btlship-session")
	gameID, ok := session.Values["gameID"].(string)
	if ok {
		http.Redirect(w, r, fmt.Sprintf("/%s", gameID), http.StatusSeeOther)
		//http.Redirect(w, r, fmt.Sprintf("/btlship/%s", gameID), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/start", http.StatusSeeOther)
		//http.Redirect(w, r, "/btlship/start", http.StatusSeeOther)
	}
}

func (app *application) startGameForm(w http.ResponseWriter, r *http.Request) {
	if len(app.gameModel.Games) == MaxGames {
		w.Write([]byte("Sorry, too many games right now. Please try after a while."))
		return
	}
	app.render(w, r, "startjoin.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

func (app *application) startGame(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := forms.New(r.PostForm)
	form.ValidateNewGameForm()

	if !form.Valid() {
		app.render(w, r, "startjoin.page.tmpl", &templateData{Form: form})
		return
	}

	pgame, err := models.NewGame(form.Values)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.gameModel.Games[pgame.ID] = pgame
	// delete game from gameModel after timeout
	go app.gameTimeout(pgame.ID)

	session, _ := app.sessionStore.Get(r, "btlship-session")
	session.Values["gameID"] = pgame.ID

	for playerID := range pgame.Players {
		session.Values["playerID"] = playerID
	}
	err = session.Save(r, w)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/%s", pgame.ID), http.StatusSeeOther)
	//http.Redirect(w, r, fmt.Sprintf("/btlship/%s", pgame.ID), http.StatusSeeOther)
}

func (app *application) playGameForm(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get(":gameid")
	pgame, _ := app.gameModel.Games[gameID]

	session, _ := app.sessionStore.Get(r, "btlship-session")

	playerID, ok := session.Values["playerID"].(string)
	if !ok {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	pgame.Mu.Lock()
	defer pgame.Mu.Unlock()
	pplayer, _ := pgame.Players[playerID]

	ptd := &templateData{
		Player: pplayer,
		Status: pgame.Status,
	}

	if pgame.Status == 1 && pgame.NextToPlay == pplayer.ID {
		ptd.Form = forms.New(nil)
		ptd.GameID = gameID
		ptd.Opponent = pgame.Players[pplayer.OpponentID].NickName
	}

	if pgame.Status == 2 {
		delete(pgame.Players, playerID)
		if len(pgame.Players) == 0 {
			delete(app.gameModel.Games, gameID)
		}
		session.Options.MaxAge = -1
		err := session.Save(r, w)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	app.render(w, r, "play.page.tmpl", ptd)
}

func (app *application) handleSse(w http.ResponseWriter, r *http.Request) {
	session, _ := app.sessionStore.Get(r, "btlship-session")
	gameID, gameOk := session.Values["gameID"].(string)
	playerID, playerOk := session.Values["playerID"].(string)

	if !gameOk || !playerOk {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	pgame, ok := app.gameModel.Games[gameID]
	if !ok {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	pplayer, ok := pgame.Players[playerID]
	if !ok {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ticker := time.NewTicker(25 * time.Second)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	for {
		select {
		//request closed
		case <-r.Context().Done():
			ticker.Stop()
			return
		case <-ticker.C:
			fmt.Fprintf(w, "data: %v\n\n", "stay alive")
			w.(http.Flusher).Flush()
		case <-pplayer.MsgChn:
			fmt.Fprintf(w, "data: %v\n\n", "refresh")
			ticker.Stop()
			return
		}
	}
}

func (app *application) joinGameForm(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get(":gameid")
	pgame, _ := app.gameModel.Games[gameID]

	ptd := &templateData{
		GameID: gameID,
		Form:   forms.New(nil),
	}

	// only 1 player in Players at this stage
	for _, pplayer := range pgame.Players {
		ptd.Opponent = pplayer.NickName
	}
	app.render(w, r, "startjoin.page.tmpl", ptd)
}

func (app *application) joinGame(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	gameID := r.URL.Query().Get(":gameid")
	pgame, _ := app.gameModel.Games[gameID]
	pgame.Mu.Lock()
	defer pgame.Mu.Unlock()
	var pplayer1 *models.Player
	// only 1 player in Players at this stage
	for _, pplayer := range pgame.Players {
		pplayer1 = pplayer
	}

	form := forms.New(r.PostForm)
	form.ValidateNewGameForm()
	if !form.Valid() {
		ptd := &templateData{
			GameID: gameID,
			Form:   form,
		}
		ptd.Opponent = pplayer1.NickName
		app.render(w, r, "startjoin.page.tmpl", ptd)
		return
	}

	pplayer2, err := models.NewPlayer(form.Values)
	if err != nil {
		app.serverError(w, err)
		return
	}
	pplayer1.OpponentID = pplayer2.ID
	pplayer2.OpponentID = pplayer1.ID
	pplayer2.StatusMsgs = []string{
		fmt.Sprintf("Waiting for %s to play.", pplayer1.NickName),
	}
	pgame.Players[pplayer2.ID] = pplayer2
	pplayer1.StatusMsgs = []string{
		fmt.Sprintf("%s has joined the game", pplayer2.NickName),
		"It's your turn to play.",
	}
	pplayer1.MsgChn <- "refresh"
	pgame.Status = 1

	session, _ := app.sessionStore.Get(r, "btlship-session")
	session.Values["gameID"] = pgame.ID
	session.Values["playerID"] = pplayer2.ID

	err = session.Save(r, w)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/%s", pgame.ID), http.StatusSeeOther)
	//http.Redirect(w, r, fmt.Sprintf("/btlship/%s", pgame.ID), http.StatusSeeOther)
}

func (app *application) playGame(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	gameID := r.URL.Query().Get(":gameid")
	pgame, _ := app.gameModel.Games[gameID]

	session, _ := app.sessionStore.Get(r, "btlship-session")
	playerID, playerOk := session.Values["playerID"].(string)
	if !playerOk {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	pgame.Mu.Lock()
	defer pgame.Mu.Unlock()

	if pgame.NextToPlay != playerID {
		http.Redirect(w, r, fmt.Sprintf("/%s", gameID), http.StatusSeeOther)
		//http.Redirect(w, r, fmt.Sprintf("/btlship/%s", gameID), http.StatusSeeOther)
		return
	}

	pplayer := pgame.Players[playerID]
	form := forms.New(r.PostForm)
	form.ValidateFireForm()
	if !form.Valid() {
		pplayer.StatusMsgs = append(pplayer.StatusMsgs, "You have entered an invalid firing position. Try again.")
		http.Redirect(w, r, fmt.Sprintf("/%s", gameID), http.StatusSeeOther)
		//http.Redirect(w, r, fmt.Sprintf("/btlship/%s", gameID), http.StatusSeeOther)
		return
	}

	field := form.Values.Get("target_pos")
	num, _ := strconv.Atoi(strings.TrimSpace(field))
	hitPos := [2]int{num / 10, num % 10}
	hitFlag := false
	shipDestroyed := ""

	pplayer.StatusMsgs = pplayer.StatusMsgs[:0]
	popponent := pgame.Players[pplayer.OpponentID]
	popponent.StatusMsgs = popponent.StatusMsgs[:0]
outer:
	for i, pship := range popponent.Ships {
		for partIndex, shipPart := range pship.Parts {
			if hitPos == shipPart.Pos {
				hitFlag = true
				delete(pship.Parts, partIndex)
				if len(pship.Parts) == 0 {
					delete(popponent.Ships, i)
					shipDestroyed = pship.Class
				}
				break outer
			}
		}
	}
	if hitFlag {
		popponent.Board[hitPos[0]][hitPos[1]] = popponent.Board[hitPos[0]][hitPos[1]] + "_fire"
		pplayer.ShotsBoard[hitPos[0]][hitPos[1]] = "hit_bomb"
		pplayer.StatusMsgs = append(pplayer.StatusMsgs, "You have HIT a ship.")
		popponent.StatusMsgs = append(popponent.StatusMsgs, "You have been hit.")
		if shipDestroyed != "" {
			pplayer.StatusMsgs = append(pplayer.StatusMsgs, "You have destroyed a "+shipDestroyed+".")
			popponent.StatusMsgs = append(popponent.StatusMsgs, "You have lost a "+shipDestroyed+".")
			if len(popponent.Ships) == 0 {
				pplayer.StatusMsgs = append(pplayer.StatusMsgs, "You have destroyed all your opponent's ships.", "You are the WINNER!")
				popponent.StatusMsgs = append(popponent.StatusMsgs, "You have lost  all your ships", "You have lost the game.")
				pgame.Status = 2
			}
		}
	} else {
		pplayer.ShotsBoard[hitPos[0]][hitPos[1]] = "splash"
		pplayer.StatusMsgs = append(pplayer.StatusMsgs, "You missed.")
		popponent.StatusMsgs = append(popponent.StatusMsgs, fmt.Sprintf("%s has missed. No casualty.", pplayer.NickName))
	}

	if pgame.Status != 2 {
		pplayer.StatusMsgs = append(pplayer.StatusMsgs, fmt.Sprintf("Waiting for %s to play.", popponent.NickName))
		popponent.StatusMsgs = append(popponent.StatusMsgs, "Your turn to play.")
		pgame.NextToPlay = popponent.ID
	}
	popponent.MsgChn <- "refresh"
	http.Redirect(w, r, fmt.Sprintf("/%s", pgame.ID), http.StatusSeeOther)
	//http.Redirect(w, r, fmt.Sprintf("/btlship/%s", pgame.ID), http.StatusSeeOther)
}
