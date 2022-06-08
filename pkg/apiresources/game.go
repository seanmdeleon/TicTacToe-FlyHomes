package apiresources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/TicTacToe-Backend/SeanDeLeon/pkg/database"
	"github.com/TicTacToe-Backend/SeanDeLeon/pkg/validator"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

/*
	RetrieveAllGames retrieves all games from the DB that are of state IN_PROGRESS

	Example Response
		{
			"error": null,
			"data": {"games": ["gameid1", "gameid2"] }
		}

	StatusCodes
	  200 Ok
	  500 InternalServerError
*/
func RetrieveAllGames(w http.ResponseWriter, r *http.Request) {

	response := Response{ErrorMessage: new(string)}
	defer json.NewEncoder(w).Encode(&response)

	games, err := dbClient.GetAllGames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		*response.ErrorMessage = err.Error()
		return
	}

	inProgressGameIds := []string{}
	for _, game := range games {
		if game.State == database.StateInProgress {
			// Keep this IN_PROGRESS game
			inProgressGameIds = append(inProgressGameIds, game.ID)
		}
	}

	response.Data = map[string]interface{}{
		"games": inProgressGameIds,
	}
	response.ErrorMessage = nil

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
		{
			"error": null,
			"data": {"gameId": "gameUUID"}
		}

	StatusCodes
	  200 Ok
	  400 BadRequest
	  500 InternalServerError
*/
func CreateNewGame(w http.ResponseWriter, r *http.Request) {

	response := Response{ErrorMessage: new(string)}
	defer json.NewEncoder(w).Encode(&response)

	type GameRequest struct {
		Players []string `json:"players" validate:"required,len=2"`
		Columns int      `json:"columns" validate:"required,eq=3"`
		Rows    int      `json:"rows" validate:"required,eq=3"`
	}

	v := validator.New()

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		*response.ErrorMessage = err.Error()
		return
	}

	gameRequest := GameRequest{}
	err = json.Unmarshal(requestBody, &gameRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		*response.ErrorMessage = err.Error()
		return
	}

	errStr := v.ValidateStruct(gameRequest)
	if errStr != nil {
		http.Error(w, *errStr, http.StatusBadRequest)
		response.ErrorMessage = errStr
		return
	}

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
		*response.ErrorMessage = "InternalServerError handling creation of new game"
		return
	}

	response.Data = map[string]interface{}{
		"gameId": id,
	}
	response.ErrorMessage = nil

	w.WriteHeader(http.StatusOK)
}

/*
	RetrieveGameState retrieves the status of a game provided the gameID

	Example Response
	{
		"error": null,
		"data":	{ "players" : ["player1", "player2"], # The list of players.
  		  		  "state": "COMPLETE/IN_PROGRESS",
           		   "winner": "player1", # IF draw, winner will be null, state will be COMPLETE.
                                # IF in progess, key should not exist.
        		}
	}
	StatusCodes
	  200 Ok
	  400 StatusBadRequest
	  404 NotFound
*/
func RetrieveGameState(w http.ResponseWriter, r *http.Request) {

	response := Response{ErrorMessage: new(string)}
	defer json.NewEncoder(w).Encode(&response)

	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		http.Error(w, "game_id not provided", http.StatusBadRequest)
		*response.ErrorMessage = "game_id not provided"
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		fmt.Println("in here!")
		*response.ErrorMessage = err.Error()
		fmt.Println("error: ", *response.ErrorMessage)
		json.NewEncoder(w).Encode(&response)
		return
	}

	response.Data = map[string]interface{}{
		"players": []string{game.Players[0], game.Players[1]},
		"state":   string(game.State),
	}

	if game.State == database.StateComplete {
		// If there is a draw, the winner is nil
		if game.Winner == nil {
			response.Data["winner"] = nil
		} else {
			// else display the winner's name
			response.Data["winner"] = game.Winner
		}
	}
	response.ErrorMessage = nil

	w.WriteHeader(http.StatusOK)
}

/*
	QuitGame quits a game by updating a game with the state of QUIT given a gameID

	Example Response
		{
			"error": null,
			"quitGame": "gameID1"
		}

	StatusCodes
	  200 Ok
	  404 NotFound
*/
func QuitGame(w http.ResponseWriter, r *http.Request) {

	response := Response{ErrorMessage: new(string)}
	defer json.NewEncoder(w).Encode(&response)

	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		http.Error(w, "game_id not provided", http.StatusBadRequest)
		*response.ErrorMessage = "game_id not provided"
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		*response.ErrorMessage = err.Error()
		return
	}

	// update the game to have a QUIT state
	game.State = database.StateQuit
	dbClient.UpdateGame(game)

	// let the UI handle the messaging
	response.Data = map[string]interface{}{
		"quitGame": gameID,
	}
	response.ErrorMessage = nil

	w.WriteHeader(http.StatusOK)
}
