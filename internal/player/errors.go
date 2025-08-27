package player

import "errors"

var (
	ErrPlayerNotFound     = errors.New("player not found")
	ErrPlayerAlreadyExists = errors.New("player already exists")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidDisplayName = errors.New("invalid display name")
)