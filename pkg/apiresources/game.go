package apiresources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/seanmdeleon/TicTacToe-FlyHomes/pkg/database"
)

/*
	RetrieveAllGames retrieves all games from the DB that are of state IN_PROGRESS

	Example Response
		{"games": ["gameid1", "gameid2"] }

	StatusCodes
	  200 Ok
	  500 InternalServerError
*/
func RetrieveAllGames(w http.ResponseWriter, r *http.Request) {

	games, err := dbClient.GetAllGames()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	inProgressGameIds := []string{}
	for _, game := range games {
		if game.State == database.StateInProgress {
			// Keep this IN_PROGRESS game
			inProgressGameIds = append(inProgressGameIds, game.ID)
		}
	}

	response := map[string][]string{
		"games": inProgressGameIds,
	}

	json.NewEncoder(w).Encode(&response)
	w.WriteHeader(http.StatusOK)
}

/*
	CreateNewGame creates a new game in the DB

	Request Body
	{
		"players": ["player1", "player2"],
		"columns": 3,
		"rows": 3
	}

	Response
		{"gameId": "gameUUID"}

	StatusCodes
	  200 Ok
	  400 BadRequest
	  500 InternalServerError
*/
func CreateNewGame(w http.ResponseWriter, r *http.Request) {

	type GameRequest struct {
		Players []string `json:"players" validate:"required,len=2"`
		Columns int      `json:"columns" validate:"required,eq=3"`
		Rows    int      `json:"rows" validate:"required,eq=3"`
	}

	v := validator.New()

	v.SetTagName("blah")

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Failed to unmarshal GameRequest json. Err: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	gameRequest := GameRequest{}
	err = json.Unmarshal(requestBody, &gameRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(gameRequest.Players) != 2 {
		http.Error(w, "There must two players listed", http.StatusBadRequest)
		return
	}

	if gameRequest.Columns != 3 {
		http.Error(w, "Number of columns must be 3", http.StatusBadRequest)
		return
	}

	if gameRequest.Rows != 3 {
		http.Error(w, "Number of rows must be 3", http.StatusBadRequest)
		return
	}

	// made it some how lol
	newBoard := [][]int{}
	for i := 0; i < gameRequest.Rows; i++ {
		row := []int{}
		for j := 0; j < gameRequest.Columns; j++ {
			row = append(row, -1)
		}
		newBoard = append(newBoard, row)
	}

	game := database.Game{
		ID:            uuid.NewV4().String(),
		Players:       map[int]string{0: gameRequest.Players[0], 1: gameRequest.Players[1]},
		Columns:       gameRequest.Columns,
		Rows:          gameRequest.Rows,
		State:         database.StateInProgress,
		Moves:         []database.Move{},
		Winner:        nil,
		NextPlayerIdx: -1,
		GameBoard:     newBoard,
	}

	id, err := dbClient.CreateNewGame(game)
	if err != nil {
		fmt.Printf("Failed to CreateNewGame in DB: %s", err.Error())
		http.Error(w, "InternalServerError handling creation of new game", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"gameId": id,
	}

	json.NewEncoder(w).Encode(&response)
	w.WriteHeader(http.StatusOK)
}

/*
	RetrieveGameState retrieves the status of a game provided the gameID

	Example Response
		{ "players" : ["player1", "player2"], # The list of players.
  		  "state": "COMPLETE/IN_PROGRESS",
           "winner": "player1", # IF draw, winner will be null, state will be COMPLETE.
                                # IF in progess, key should not exist.
        }

	StatusCodes
	  200 Ok
	  400 StatusBadRequest
	  404 NotFound
*/
func RetrieveGameState(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	type response struct {
		Players []string `json:"players"`
		State   string   `json:"state"`
		Winner  *string  `json:"winner,omitempty"`
	}

	resp := response{
		Players: []string{game.Players[0], game.Players[1]},
		State:   string(game.State),
	}

	if game.State == database.StateComplete {
		// If there is a draw, the winner is nil
		if game.Winner == nil {
			resp.Winner = nil
		} else {
			// else display the winner's name
			resp.Winner = game.Winner
		}
	}

	json.NewEncoder(w).Encode(&resp)
	w.WriteHeader(http.StatusOK)
}

/*
	QuitGame quits a game by updating a game with the state of QUIT given a gameID

	Example Response
		{"quitGame": "gameID1" }

	StatusCodes
	  200 Ok
	  404 NotFound
*/
func QuitGame(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// update the game to have a QUIT state
	game.State = database.StateQuit
	dbClient.UpdateGame(game)

	// let the UI handle the messaging
	response := map[string]string{
		"quitGame": gameID,
	}

	json.NewEncoder(w).Encode(&response)
	w.WriteHeader(http.StatusOK)
}
