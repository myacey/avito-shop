package postgresrepo

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/repository"
)

type PostgresStoreRepo struct {
	store db.Querier
}

func NewPostgresStoreRepo(store db.Querier) repository.StoreRepository {
	return &PostgresStoreRepo{store}
}

func (r *PostgresStoreRepo) GetItemInfo(c context.Context, itemName string) (*db.Item, error) {
	item, err := r.store.GetItemFromStore(c, itemName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrInvalidItemName
		}
		return nil, err
	}

	return &item, nil
}
