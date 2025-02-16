package repository

import (
	"context"
	"errors"

	db "github.com/myacey/avito-shop/db/sqlc"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository interface {
	CreateUser(c context.Context, username, password string) (*db.User, error)
	GetUser(c context.Context, username string) (*db.User, error)

	GetUserForUpdate(c context.Context, username string) (*db.User, error)
	UpdateBalance(c context.Context, userID int32, newCointCount int32) (*db.User, error)
	UpdateTwoUsersBalance(c context.Context, fromUsername, toUsername string, coinsAmount int32) ([]*db.User, error)
}
