package database

import (
	"fmt"
)

/*
	This is an extremely simple and super restricted "in membory database"
	I am using this apporach to simplify this coding challenge
	Originally my plan was to use awsdynamodb and setup the table in dynamoDB under my own AWS account.
	Instead I've decided to use a simple map to store the data while the program is running. I've chosen to talk about expanding the project
	to full scale in the follow up interview
*/

// Custom typing for some known string values
type MoveType string
type State string

/*
	ticTacToeDBTable is the structure that represents a database table
	The structure is a map[gameID] -> Game, where gameID is the PK of the table
	Game holds all of the information about a given game
	Essentially, this table is just a uuid pointing to a game configuration
	The table is lowercase so that only this file is aware of it
	This table is acts as an InMemory DB
*/
var ticTacToeDbTable map[string]Game

const (
	MoveTypeMove MoveType = "MOVE"
	MoveTypeQuit MoveType = "QUIT"

	StateComplete   State = "COMPLETE"
	StateInProgress State = "IN_PROGRESS"
	StateQuit       State = "QUIT"
)

// DB is the interface that holds the methods for accessing the TicTacToe DB
type DB interface {
	GetGameWithID(id string) (Game, error)
	GetAllGames() ([]Game, error)
	CreateNewGame(game Game) (string, error)
	UpdateGame(game Game) error
}

// Client is the client the implements the DB interface. The holds access to the InMemory ticTacToeDBTable
type Client struct {

	// a channel for locking the ticTacToeDBTable to simulate atomic read and writes
	channelLock chan bool
}

// Verifying if the Client struct is indeed implementing the DB interface
var _ DB = (*Client)(nil)

// Game represents the configuration of a TicTacToe Game
type Game struct {
	ID            string         `json:"id"`
	Players       map[int]string `json:"players"`
	Columns       int            `json:"columns"`
	Rows          int            `json:"rows"`
	State         State          `json:"state"`
	Winner        *string        `json:"winner"`
	Moves         []Move         `json:"moves"`
	NextPlayerIdx int            `json:"nextPlayerIdx"` // The index into the Player array of the next move, either 0 or 1
	GameBoard     [][]int        `json:"gameBoard"`     // The game board
}

// Move represents data about a TicTacToe move
type Move struct {
	Type   MoveType `json:"type"`
	Player string   `json:"player"`
	Row    int      `json:"row"`
	Col    int      `json:"col"`
}

// New returns a new Client to access the DB
func New() *Client {

	// initialize the InMemory DB table
	ticTacToeDbTable = map[string]Game{}

	// initialize the channel lock
	c := make(chan bool, 1)
	c <- true
	return &Client{
		channelLock: c,
	}
}

// GetGameWithID  returns a game from the DB provided the game id
// return an error if no game with the provided id exists
func (c *Client) GetGameWithID(id string) (Game, error) {

	<-c.channelLock
	defer func() { c.channelLock <- true }()

	game, ok := ticTacToeDbTable[id]
	if !ok {
		return Game{}, fmt.Errorf("No game exists with provided game_id %s", id)
	}

	return game, nil
}

// GetAllGames returns a list of all TicTacToe games listed in the DB table
func (c *Client) GetAllGames() ([]Game, error) {

	result := []Game{}

	<-c.channelLock
	defer func() { c.channelLock <- true }()

	for _, game := range ticTacToeDbTable {
		result = append(result, game)
	}

	return result, nil
}

// CreateNewGame creates a new game submitted by the caller as a row in the DB, return the gameID provided
func (c *Client) CreateNewGame(game Game) (string, error) {

	<-c.channelLock
	defer func() { c.channelLock <- true }()

	ticTacToeDbTable[game.ID] = game

	return game.ID, nil
}

// UpdateGame updates an existing game
func (c *Client) UpdateGame(game Game) error {

	<-c.channelLock
	defer func() { c.channelLock <- true }()

	_, ok := ticTacToeDbTable[game.ID]

	if !ok {
		// This state should never be reached since the caller SHOULD call the GetGameWithID method first
		return fmt.Errorf("Failed to Update game. Game with game_id %s does not exist", game.ID)
	}

	ticTacToeDbTable[game.ID] = game
	return nil
}
