package apiresources

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/seanmdeleon/TicTacToe-FlyHomes/pkg/database"
)

// make this package variable an interface to enable mocked testing
var dbClient database.DB

// GetRouter builds the main router with the tictactoe subrouter
func GetRouter() *mux.Router {

	mainRouter := mux.NewRouter().StrictSlash(true)

	subRouter := mainRouter.PathPrefix("/tictactoe").Subrouter()
	subRouter.HandleFunc("", RetrieveAllGames).Name("RetrieveAllGames").Methods("GET")
	subRouter.HandleFunc("", CreateNewGame).Name("CreateNewGame").Methods("POST")
	subRouter.HandleFunc("/{game_id}", RetrieveGameState).Name("RetrieveGameState").Methods("GET")
	subRouter.HandleFunc("/{game_id}/moves", RetrieveListOfMoves).Name("RetrieveListOfMoves").Methods("GET")
	subRouter.HandleFunc("/{game_id}/{player_id}", PostAMove).Name("PostAMove").Methods("POST")
	subRouter.HandleFunc("/{game_id}/moves/{move_number}", RetrieveAMove).Name("RetrieveAMove").Methods("GET")
	subRouter.HandleFunc("/{game_id}/quit", QuitGame).Name("QuitGame").Methods("PUT")

	// assign the package DB client
	GetNewDBClient()

	return mainRouter
}

// CaselessMatcher forces URL paths to lowercase so they are matched invariant of their input casing.
// Subsequent routes must be updated to use lowercase, or they will not be matched.
func CaselessMatcher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.ToLower(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// GetNewDBClient initializes the DB client at the top of this file which can be used by the entire apiresources package
func GetNewDBClient() {
	dbClient = database.New()
}
