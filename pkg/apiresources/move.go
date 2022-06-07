package apiresources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/seanmdeleon/TicTacToe-FlyHomes/pkg/database"
)

/*
	RetrieveListOfMoves the list or sublist of moves from a given game_id
	Optional query arguments to retrieve only a a sublist
	If the optionally provided 'start' is out of range, return a 404
	If the optionally provided 'until' is out of range, simply return start -> last of all moves
	'start' defaults to 0 and 'until' defaults to (the total number of moves - 1)

	Example Query
		GET /tictactoe/{game_id}/moves?start=0&until=1

	Example Response
		{
  			"moves": [{"type": "MOVE", "player": "player1", "row":1, "column":1 }, {"type": "QUIT", "player": "player2"}]
		}

	StatusCodes
	  200 Ok
	  400 BadRequest
	  404 NotFound
	  500 InternalServerError
*/
func RetrieveListOfMoves(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("No gameID provided")
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil {
		fmt.Printf("Failed to find game with gameID %s. Err: %s\n", gameID, err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// default values for start and until
	start := 0
	until := len(game.Moves) - 1

	// optional query params
	startStr := r.URL.Query().Get("start")
	untilStr := r.URL.Query().Get("until")

	// check if startStr exists and if it's actually an integer
	if len(startStr) > 0 {
		start, err = strconv.Atoi(startStr)
		if err != nil {
			// append error
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// check if untilStr exists and if it's actually an integer
	if len(untilStr) > 0 {
		until, err = strconv.Atoi(untilStr)
		if err != nil {
			// append error
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// take the lowest most valid until value quietly
		if until >= len(game.Moves) {
			until = len(game.Moves) - 1
		}
	}

	ok, err = validateStartAndUntilValues(start, until, len(game.Moves))
	if !ok {
		fmt.Printf("Err: %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := map[string][]database.Move{
		"moves": game.Moves[start : until+1],
	}

	json.NewEncoder(w).Encode(&response)
	w.WriteHeader(http.StatusOK)
}

func validateStartAndUntilValues(start, until, totalNumMoves int) (bool, error) {

	if start > until {
		return false, fmt.Errorf("start must be <= until")
	}

	if start >= totalNumMoves {
		return false, fmt.Errorf("This game has a total of %d moves, so start must be < %d", totalNumMoves, totalNumMoves)
	}

	return true, nil
}

/*
	RetrieveAMove returns a move provided a move_number and a game_id
	move_number is 0 offset, provided by the POST move endpoint

	Example Response
		{
			"type" : "MOVE",
			"player": "1", # number corresponding to the player
			"row": "1",
			"column": 2
		}

	StatusCodes
	  200 Ok
	  400 BadRequest
	  404 NotFound
	  500 InternalServerError
*/
func RetrieveAMove(w http.ResponseWriter, r *http.Request) {

	// Retrieve game_id and move_number and validate them
	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("No game_id provided")
		return
	}

	moveNumberStr, ok := vars["move_number"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("No move_number provided")
		return
	}
	moveNumber, err := strconv.Atoi(moveNumberStr)
	if err != nil {
		fmt.Printf("move_number must be an integer")
		// append error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil {
		fmt.Printf("Failed to find game with gameID %s. Err: %s\n", gameID, err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// move_number must be within range, and is 0 offset
	if moveNumber >= len(game.Moves) {
		fmt.Printf("move_number %d is out of range\n", moveNumber)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := game.Moves[moveNumber]
	json.NewEncoder(w).Encode(&response)
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
			"move": "{gameId}/moves/{move_number}"
			"winner": "player1" // omitempty
		}

	StatusCodes
	  200 Ok
	  400 BadRequest
	  404 NotFound
	  409 NotPlayersTurn
	  500 InternalServerError
*/
func PostAMove(w http.ResponseWriter, r *http.Request) {

	type MoveRequest struct {
		Column int `json:"column" validate:"required"`
		Row    int `json:"row" validate:"required"`
	}

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	moveRequest := MoveRequest{}
	err = json.Unmarshal(requestBody, &moveRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// validate!!! but only for user input

	// Retrieve game_id and move_number and validate them
	vars := mux.Vars(r)
	gameID, ok := vars["game_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("No game_id provided")
		return
	}

	playerIDStr, ok := vars["player_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("No player_id provided")
		return
	}
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		fmt.Printf("player_id must be an integer")
		// append error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	game, err := dbClient.GetGameWithID(gameID)
	if err != nil {
		fmt.Printf("Failed to find game with gameID %s. Err: %s\n", gameID, err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Player not found
	if _, ok := game.Players[playerID]; !ok {
		fmt.Printf("Player with playerID %d is not found\n", playerID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Not the current player's turn
	if game.NextPlayerIdx != -1 && game.NextPlayerIdx != playerID {
		fmt.Printf("Is is not player %d's turn\n", playerID)
		w.WriteHeader(http.StatusConflict)
		return
	}

	response := map[string]string{}
	moveNumber, err := playMove(moveRequest.Row, moveRequest.Column, playerID, &game)
	if err != nil {
		fmt.Printf("Failed to play the move, it is illegal. %s\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response["move"] = fmt.Sprintf("%s/moves/%d", gameID, moveNumber)

	// Store next player
	if game.NextPlayerIdx == 0 {
		game.NextPlayerIdx = 1
	} else {
		game.NextPlayerIdx = 0
	}

	// check the board for a winner and store the winner in the response
	if checkBoardForWinner(moveRequest.Row, moveRequest.Column, playerID, &game) {
		fmt.Printf("Winner! player: %s\n", game.Players[playerID])
		game.State = database.StateComplete
		response["winner"] = game.Players[playerID]
	}

	// Update the move in the DB
	err = dbClient.UpdateGame(game)
	if err != nil {
		fmt.Printf("Failed to update the game in the DB. Err: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&response)
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
