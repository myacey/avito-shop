package repository

import (
	"context"
	"errors"

	db "github.com/myacey/avito-shop/db/sqlc"
)

var ErrInvalidItemName = errors.New("no item with this name")

type StoreRepository interface {
	GetItemInfo(c context.Context, itemName string) (*db.Item, error)
}
