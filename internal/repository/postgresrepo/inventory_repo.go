package postgresrepo

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/repository"
)

type PostgresInventoryRepo struct {
	store db.Querier
}

func NewPostgresInventoryRepo(store db.Querier) repository.InventoryRepository {
	return &PostgresInventoryRepo{store}
}

func (r *PostgresInventoryRepo) AddItemToInventory(c context.Context, userID int32, itemType string) error {
	arg := db.BuyItemParams{
		UserID:   userID,
		ItemType: itemType,
	}
	err := r.store.BuyItem(c, arg)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresInventoryRepo) GetInventory(c context.Context, userID int32) ([]*db.Inventory, error) {
	inv, err := r.store.GetInventory(c, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNoInventoryItems
		}
		return nil, err
	}

	ans := make([]*db.Inventory, len(inv))
	for i := range inv {
		ans[i] = &inv[i]
	}

	return ans, nil
}
