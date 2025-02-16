package repository

import (
	"context"
	"errors"

	db "github.com/myacey/avito-shop/db/sqlc"
)

var ErrNoTransfers = errors.New("no transafers found")

type TransferRepository interface {
	CreateMoneyTransfer(c context.Context, fromUsername, toUsername string, amount int32) (*db.Transfer, error)
	GetTransfersWithUser(c context.Context, username string) ([]*db.Transfer, error)
}
