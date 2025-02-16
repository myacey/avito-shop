package repository

import (
	"context"
	"errors"

	db "github.com/myacey/avito-shop/db/sqlc"
)

var ErrNoInventoryItems = errors.New("empty inventory")

type InventoryRepository interface {
	AddItemToInventory(c context.Context, userID int32, itemType string) error
	GetInventory(c context.Context, userID int32) ([]*db.Inventory, error)
}
