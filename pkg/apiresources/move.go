package apiresources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/TicTacToe-Backend/SeanDeLeon/pkg/database"
	"github.com/TicTacToe-Backend/SeanDeLeon/pkg/validator"
	"github.com/gorilla/mux"
)

/*
	RetrieveListOfMoves the list or sublist of moves from a given game_id
	Optional query arguments to retrieve only a a sublist
	If the optionally provided 'start' is out of range, default to 0
	If the optionally provided 'until' is out of range, simply return start -> last of all moves
	'start' defaults to 0 and 'until' defaults to (the total number of moves - 1)

	Example Query
		GET /tictactoe/{game_id}/moves?start=0&until=1

	Example Response
		{
			"error": null,
			"data": {
  				"moves": [{"type": "MOVE", "player": "player1", "row":1, "column":1 }, {"type": "QUIT", "player": "player2"}]
			}
		}

	StatusCodes
	  200 Ok
	  400 BadRequest
	  404 NotFound
	  500 InternalServerError
*/
func RetrieveListOfMoves(w http.ResponseWriter, r *http.Request) {

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
		fmt.Printf("Failed to find game with gameID %s. Err: %s\n", gameID, err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		*response.ErrorMessage = err.Error()
		return
	}

	// default values for start and until
	start := 0
	until := len(game.Moves) - 1

	// optional query params
	startStr := r.URL.Query().Get("start")
	untilStr := r.URL.Query().Get("until")

	// check if startStr exists
	if len(startStr) > 0 {
		// no need to error check, if it's wrong then just default to 0
		start, _ = strconv.Atoi(startStr)
	}

	// check if untilStr exists
	if len(untilStr) > 0 {
		until, _ = strconv.Atoi(untilStr)
		// take the lowest most valid until value quietly
		if until >= len(game.Moves) {
			until = len(game.Moves) - 1
		}
	}

	ok, errMsgs := validateStartAndUntilValues(start, until, len(game.Moves))
	if !ok {
		fmt.Printf("Error: %s\n", errMsgs)
		http.Error(w, errMsgs, http.StatusBadRequest)
		*response.ErrorMessage = errMsgs
		return
	}

	response.Data = map[string]interface{}{
		"moves": game.Moves[start : until+1],
	}
	response.ErrorMessage = nil

	w.WriteHeader(http.StatusOK)
}

func validateStartAndUntilValues(start, until, totalNumMoves int) (bool, string) {

	valid := true
	errMsgs := ""
	if totalNumMoves == 0 {
		return false, fmt.Sprintf("There are no moves for this game")
	}

	if start > until {
		valid = false
		errMsgs += fmt.Sprintf("'start' must be less than or equal to  'until'. ")
	}

	if start >= totalNumMoves {
		valid = false
		errMsgs += fmt.Sprintf("This game has a total of %d moves, so start must be less than %d.", totalNumMoves, totalNumMoves)
	}

	return valid, errMsgs
}

/*
	RetrieveAMove returns a move provided a move_number and a game_id
	move_number is 0 offset, provided by the POST move endpoint

	Example Response
		{
			"error": null,
			"data" : {
				"type" : "MOVE",
				"player": "1", # number corresponding to the player
				"row": "1",
				"column": 2
			}
		}

	StatusCodes
	  200 Ok
	  400 BadRequest
	  404 NotFound
	  500 InternalServerError
*/
func RetrieveAMove(w http.ResponseWriter, r *http.Request) {

	response := Response{ErrorMessage: new(string)}
	defer json.NewEncoder(w).Encode(&response)

	// Retrieve game_id and move_number and validate them
	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		http.Error(w, "game_id not provided", http.StatusBadRequest)
		*response.ErrorMessage = "game_id not provided"
		return
	}

	moveNumberStr, ok := vars["move_number"]
	if !ok {
		http.Error(w, "move_number not provided", http.StatusBadRequest)
		*response.ErrorMessage = "move_number not provided"
		return
	}

	moveNumber, err := strconv.Atoi(moveNumberStr)
	if err != nil {
		http.Error(w, "move_number must be an integer", http.StatusBadRequest)
		*response.ErrorMessage = "move_number must be an integer"
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil {
		fmt.Printf("Failed to find game with gameID %s. Err: %s\n", gameID, err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		*response.ErrorMessage = err.Error()
		return
	}

	// move_number must be within range, and is 0 offset
	if moveNumber >= len(game.Moves) {
		http.Error(w, fmt.Sprintf("move_number %d is out of range\n", moveNumber), http.StatusBadRequest)
		*response.ErrorMessage = fmt.Sprintf("move_number %d is out of range\n", moveNumber)
		return
	}

	response.Data = map[string]interface{}{
		"move": game.Moves[moveNumber],
	}
	response.ErrorMessage = nil

	w.WriteHeader(http.StatusOK)
}

/*
	PostAMove posts a move to the current game provided a game_id and player_id
	player_id is just either 0 or 1

	POST /tictactoe/{game_id}/{player_id}

	Example Request
		{
			"row" : 1,
			"column" : 1
		}

	Example Response
		{
			"error": null,
			"data" : {
				"move": "{gameId}/moves/{move_number}"
				"winner": "player1" // omitempty
			}
		}

	StatusCodes
	  200 Ok
	  400 BadRequest
	  404 NotFound
	  409 NotPlayersTurn
	  500 InternalServerError
*/
func PostAMove(w http.ResponseWriter, r *http.Request) {

	response := Response{ErrorMessage: new(string)}
	defer json.NewEncoder(w).Encode(&response)

	type MoveRequest struct {
		Column *int `json:"column" validate:"required,lte=2,gte=0"`
		Row    *int `json:"row" validate:"required,lte=2,gte=0"`
	}

	v := validator.New()

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		*response.ErrorMessage = err.Error()
		return
	}

	moveRequest := MoveRequest{}
	err = json.Unmarshal(requestBody, &moveRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		*response.ErrorMessage = fmt.Sprintf("Failed to unmarshal data, moveRequest is malformed. %s", err.Error())
		return
	}

	// Retrieve game_id and player_id and validate them
	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		http.Error(w, "game_id not provided", http.StatusBadRequest)
		*response.ErrorMessage = "game_id not provided"
		return
	}

	playerIDStr, ok := vars["player_id"]
	if !ok {
		http.Error(w, "player_id not provided", http.StatusBadRequest)
		*response.ErrorMessage = "player_id not provided"
		return
	}
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		http.Error(w, "player_id must be an integer", http.StatusBadRequest)
		*response.ErrorMessage = "player_id must be an integer"
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil || game.State == database.StateComplete || game.State == database.StateQuit {
		fmt.Printf("Failed to find game with gameID %s. Err: %s\n", gameID, err.Error())
		http.Error(w, fmt.Sprintf("Failed to find an IN_PROGRESS game with this gameID. %s", err.Error()), http.StatusNotFound)
		*response.ErrorMessage = err.Error()
		return
	}

	// Player not found
	if _, ok := game.Players[playerID]; !ok {
		e := fmt.Errorf("Player with playerID %d is not found\n", playerID)
		http.Error(w, e.Error(), http.StatusNotFound)
		*response.ErrorMessage = e.Error()
		return
	}

	// Not the current player's turn
	if game.NextPlayerIdx != -1 && game.NextPlayerIdx != playerID {
		e := fmt.Errorf("Is is not player %d's turn\n", playerID)
		http.Error(w, e.Error(), http.StatusNotFound)
		*response.ErrorMessage = e.Error()
		return
	}

	errStr := v.ValidateStruct(moveRequest)
	if errStr != nil {
		http.Error(w, *errStr, http.StatusBadRequest)
		response.ErrorMessage = errStr
		return
	}

	moveNumber, err := playMove(*moveRequest.Row, *moveRequest.Column, playerID, &game)
	if err != nil {
		e := fmt.Errorf("Failed to play the move, it is illegal. %s\n", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		*response.ErrorMessage = e.Error()
		return
	}

	response.Data = map[string]interface{}{
		"move": fmt.Sprintf("%s/moves/%d", gameID, moveNumber),
	}

	// Store next player
	if game.NextPlayerIdx == -1 {
		if playerID == 0 {
			game.NextPlayerIdx = 1
		} else {
			game.NextPlayerIdx = 0
		}
	} else if game.NextPlayerIdx == 0 {
		game.NextPlayerIdx = 1
	} else {
		game.NextPlayerIdx = 0
	}

	// check the board for a winner and store the winner in the response
	if checkBoardForWinner(*moveRequest.Row, *moveRequest.Column, playerID, &game) {
		fmt.Printf("Winner! player: %s\n", game.Players[playerID])
		game.State = database.StateComplete
		response.Data["winner"] = game.Players[playerID]
	} else {
		// There is no winner, if the number of moves = 9, then we know we have a DRAW and there is no winner
		// NOTE this is for a strict 3x3 board only. The real check would be against game.Rows*game.Columns
		if len(game.Moves) == 9 {
			// winner remains null
			game.State = database.StateComplete
		}
	}

	// Update the move in the DB
	err = dbClient.UpdateGame(game)
	if err != nil {
		e := fmt.Errorf("Failed to update the game in the DB. %s\n", err.Error())
		http.Error(w, e.Error(), http.StatusInternalServerError)
		*response.ErrorMessage = e.Error()
		return
	}

	response.ErrorMessage = nil
	w.WriteHeader(http.StatusOK)
}

// try to play the move, return a moveNumber and/or and error
func playMove(row, col, playerID int, game *database.Game) (int, error) {

	if row > game.Rows || row < 0 {
		return -1, fmt.Errorf("row provided (%d) is out of range [0-%d]", row, game.Rows-1)
	}

	if col > game.Columns || col < 0 {
		return -1, fmt.Errorf("col provided (%d) is out of range [0-%d]", col, game.Columns-1)
	}

	if game.GameBoard[row][col] != -1 {
		return -1, fmt.Errorf("move with row %d and col %d is already taken", row, col)
	}

	// Assign the square to the playerID
	game.GameBoard[row][col] = playerID

	// make note of the move
	game.Moves = append(game.Moves, database.Move{
		Type:   database.MoveTypeMove,
		Player: game.Players[playerID],
		Row:    row,
		Col:    col,
	})

	// return the move number, which is offset by 0
	return len(game.Moves) - 1, nil
}

func checkBoardForWinner(row, col, playerID int, game *database.Game) bool {

	// Check Left to Right
	squareCount := game.Columns
	colIdx := col
	for i := 1; i <= game.Columns; i++ {
		colIdx = colIdx % game.Columns
		if game.GameBoard[row][colIdx] != playerID {
			break
		}
		squareCount--
		colIdx++
	}

	if squareCount == 0 {
		// found winner!
		return true
	}

	// Check Up and Down
	squareCount = game.Rows
	rowIdx := row
	for i := 1; i <= game.Rows; i++ {
		rowIdx = rowIdx % game.Rows
		if game.GameBoard[rowIdx][col] != playerID {
			break
		}
		squareCount--
		rowIdx++
	}

	if squareCount == 0 {
		// found winner!
		return true
	}

	// TopLeft to BottomRight Diagonal
	squareCount = game.Rows
	r := 0
	c := 0
	for r < game.Rows {
		if game.GameBoard[r][c] != playerID {
			break
		}
		squareCount--
		r++
		c++
	}

	if squareCount == 0 {
		// found winner!
		return true
	}

	// TopRight to BottomLeft Diagonal
	squareCount = game.Rows
	r = 0
	c = game.Columns - 1
	for r < game.Rows {
		if game.GameBoard[r][c] != playerID {
			break
		}
		squareCount--
		r++
		c--
	}

	if squareCount == 0 {
		// found winner!
		return true
	}

	return false
}
