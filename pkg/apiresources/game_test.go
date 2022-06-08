package apiresources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/seanmdeleon/TicTacToe-FlyHomes/mocks"
	"github.com/seanmdeleon/TicTacToe-FlyHomes/pkg/database"
	"github.com/stretchr/testify/assert"
)

// A test file for only game.go
// I only had time to write a few test cases. In real life I take unittesting seriously

func TestRetrieveAllGamesNoGamesSuccess(t *testing.T) {

	// Setup the DB mock and assign it to the package dBClient interface
	dbMock := mocks.DB{}
	dbClient = &dbMock

	// The request
	r := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: "/tictactoe",
		},
	}

	// This implements the http.ResponseWriter interface
	w := httptest.NewRecorder()

	response := Response{}

	// First test no games created yet
	dbMock.On("GetAllGames").Return([]database.Game{}, nil)
	RetrieveAllGames(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 1, len(response.Data))
	assert.Equal(t, []interface{}{}, response.Data["games"])
	assert.Nil(t, response.ErrorMessage)
}

func TestRetrieveAllGamesManyGamesSuccess(t *testing.T) {

	// Setup the DB mock and assign it to the package dBClient interface
	dbMock := mocks.DB{}
	dbClient = &dbMock

	// The request
	r := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: "/tictactoe",
		},
	}

	// This implements the http.ResponseWriter interface
	w := httptest.NewRecorder()

	response := Response{}

	// First test no games created yet
	dbMock.On("GetAllGames").Return(generateGames(), nil)
	RetrieveAllGames(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	_ = json.Unmarshal(w.Body.Bytes(), &response)
	inProgGameIds := response.Data["games"].([]interface{})

	assert.Equal(t, 2, len(inProgGameIds))
	assert.Equal(t, []interface{}{"gameID1", "gameID2"}, inProgGameIds)
	assert.Nil(t, response.ErrorMessage)
}

func TestRetrieveGameStateFailure(t *testing.T) {

	// Setup the DB mock and assign it to the package dBClient interface
	dbMock := mocks.DB{}
	dbClient = &dbMock

	// The request
	r := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: "/tictactoe/gameID172",
		},
	}

	r = mux.SetURLVars(r, map[string]string{
		"game_id": "gameID172",
	})

	// This implements the http.ResponseWriter interface
	w := httptest.NewRecorder()

	// First test no games created yet
	dbMock.On("GetGameWithID", "gameID172").Return(database.Game{}, fmt.Errorf("No game exists with provided game_id %s", "gameID172"))
	RetrieveGameState(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func generateGames() []database.Game {
	newBoard := [][]int{}
	for i := 0; i < 3; i++ {
		row := []int{}
		for j := 0; j < 3; j++ {
			row = append(row, -1)
		}
		newBoard = append(newBoard, row)
	}

	return []database.Game{
		{
			ID:            "gameID1",
			Players:       map[int]string{0: "player1", 1: "player2"},
			Columns:       3,
			Rows:          3,
			State:         database.StateInProgress,
			NextPlayerIdx: -1,
			GameBoard:     newBoard,
		},
		{
			ID:            "gameID2",
			Players:       map[int]string{0: "player1", 1: "player2"},
			Columns:       3,
			Rows:          3,
			State:         database.StateInProgress,
			NextPlayerIdx: -1,
			GameBoard:     newBoard,
		},
		{
			ID:            "gameID3",
			Players:       map[int]string{0: "player1", 1: "player2"},
			Columns:       3,
			Rows:          3,
			State:         database.StateQuit,
			NextPlayerIdx: -1,
			GameBoard:     newBoard,
		},
		{
			ID:            "gameID4",
			Players:       map[int]string{0: "player1", 1: "player2"},
			Columns:       3,
			Rows:          3,
			State:         database.StateComplete,
			NextPlayerIdx: 1,
			GameBoard:     newBoard,
		},
	}
}
