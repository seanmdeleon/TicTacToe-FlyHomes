// Code generated by mockery v2.2.1. DO NOT EDIT.

package mocks

import (
	database "github.com/seanmdeleon/TicTacToe-FlyHomes/pkg/database"
	mock "github.com/stretchr/testify/mock"
)

// DB is an autogenerated mock type for the DB type
type DB struct {
	mock.Mock
}

// CreateNewGame provides a mock function with given fields: game
func (_m *DB) CreateNewGame(game database.Game) (string, error) {
	ret := _m.Called(game)

	var r0 string
	if rf, ok := ret.Get(0).(func(database.Game) string); ok {
		r0 = rf(game)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(database.Game) error); ok {
		r1 = rf(game)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllGames provides a mock function with given fields:
func (_m *DB) GetAllGames() ([]database.Game, error) {
	ret := _m.Called()

	var r0 []database.Game
	if rf, ok := ret.Get(0).(func() []database.Game); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]database.Game)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGameWithID provides a mock function with given fields: id
func (_m *DB) GetGameWithID(id string) (database.Game, error) {
	ret := _m.Called(id)

	var r0 database.Game
	if rf, ok := ret.Get(0).(func(string) database.Game); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Get(0).(database.Game)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateGame provides a mock function with given fields: game
func (_m *DB) UpdateGame(game database.Game) error {
	ret := _m.Called(game)

	var r0 error
	if rf, ok := ret.Get(0).(func(database.Game) error); ok {
		r0 = rf(game)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}