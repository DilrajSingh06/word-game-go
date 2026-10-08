package game

import "errors"

var (
	ErrWrongLength = errors.New("guess must be exactly 5 letters")
	ErrNotLetters  = errors.New("guess must contain only letters")
	ErrNotAWord    = errors.New("not in the word list")
	ErrGameOver    = errors.New("this game is already finished")
	ErrOutOfTries  = errors.New("no guesses left")
)
