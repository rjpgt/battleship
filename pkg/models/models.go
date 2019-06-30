package models

import (
	"crypto/rand"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// ShipPart is made of a location, Pos,
//+ and an image for that location
type ShipPart struct {
	Pos [2]int
	Img string
}

// ShipT is the battleship type
type ShipT struct {
	Class string
	Parts map[int]ShipPart
}

func NewShip(class string, field string) *ShipT {
	squares := strings.Split(field, ",")
	nums := make([]int, len(squares))
	posns := make([][2]int, len(squares))
	for i := range squares {
		nums[i], _ = strconv.Atoi(strings.TrimSpace(squares[i]))
		posns[i] = [2]int{nums[i] / 10, nums[i] % 10}
	}
	rowImageNames := [3]string{"end_left", "mid_h", "end_right"}
	colImageNames := [3]string{"end_top", "mid_v", "end_bottom"}
	var imageNames [3]string
	if posns[0][0] == posns[1][0] {
		imageNames = rowImageNames
	} else {
		imageNames = colImageNames
	}
	parts := map[int]ShipPart{}
	parts[0] = ShipPart{Pos: posns[0], Img: imageNames[0]}
	parts[len(posns)-1] = ShipPart{Pos: posns[len(posns)-1], Img: imageNames[2]}
	for i := 1; i < len(posns)-1; i++ {
		parts[i] = ShipPart{Pos: posns[i], Img: imageNames[1]}
	}

	return &ShipT{Class: class, Parts: parts}
}

// Player represents a battleship game player
type Player struct {
	// array zero value is not nil unlike that of slice
	//+ so need not explicitly initialize it.
	Board      [10][10]string
	FlashMsg   string
	ID         string
	MsgChn     chan string
	NickName   string
	OpponentID string
	//Ships      [5]ShipT
	Ships      map[int]*ShipT
	Shots      [][2]int
	ShotsBoard [10][10]string
	StatusMsgs []string
}

func NewPlayer(formFields url.Values) (*Player, error) {
	id, err := fakeUUID()
	if err != nil {
		return nil, err
	}
	btlship := NewShip("battleship", formFields.Get("btlship"))
	cruiser := NewShip("cruiser", formFields.Get("cruiser"))
	frigate := NewShip("frigate", formFields.Get("frigate"))
	destroyer := NewShip("destroyer", formFields.Get("destroyer"))
	patrolboat := NewShip("patrolboat", formFields.Get("patrolboat"))

	player := Player{
		ID:       id,
		MsgChn:   make(chan string, 1),
		NickName: formFields.Get("username"),
		Ships:    map[int]*ShipT{0: btlship, 1: cruiser, 2: frigate, 3: destroyer, 4: patrolboat},
		Shots:    [][2]int{},
	}
	for _, pship := range player.Ships {
		for _, shipPart := range pship.Parts {
			posn := shipPart.Pos
			player.Board[posn[0]][posn[1]] = shipPart.Img
		}
	}

	return &player, nil
}

// Game represents a battleship game
type Game struct {
	ID         string
	Mu         sync.Mutex
	NextToPlay string
	Players    map[string]*Player
	Status     int //0 - starting, 1 - playing, 2 - ended
}

func NewGame(formFields url.Values) (*Game, error) {
	id, err := fakeUUID()
	if err != nil {
		return nil, err
	}
	pplayer, err := NewPlayer(formFields)
	if err != nil {
		return nil, err
	}
	//joinURL := fmt.Sprintf("https://amrtam.in/btlship/join/%s", id)
	joinURL := fmt.Sprintf("/join/%s", id)
	pplayer.StatusMsgs = []string{
		fmt.Sprintf("Invite opponent to %s.", joinURL),
		"Waiting for opponent to join.",
	}
	game := Game{
		ID:         id,
		Players:    map[string]*Player{},
		NextToPlay: pplayer.ID,
		Status:     0,
	}
	game.Players[pplayer.ID] = pplayer
	return &game, nil
}

// GameModel stores all games
type GameModel struct {
	Games map[string]*Game
}

func fakeUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
