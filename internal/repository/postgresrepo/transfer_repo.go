package postgresrepo

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/repository"
)

type PostgresTransferRepo struct {
	store db.Querier
}

func NewPostgresTransferRepo(store db.Querier) repository.TransferRepository {
	return &PostgresTransferRepo{store}
}

func (r *PostgresTransferRepo) CreateMoneyTransfer(c context.Context, fromUsername, toUsername string, amount int32) (*db.Transfer, error) {
	arg := db.CreateMoneyTransferParams{
		FromUsername: fromUsername,
		ToUsername:   toUsername,
		Amount:       amount,
	}

	transfer, err := r.store.CreateMoneyTransfer(c, arg)
	if err != nil {
		return nil, err
	}

	return &transfer, nil
}

func (r *PostgresTransferRepo) GetTransfersWithUser(c context.Context, username string) ([]*db.Transfer, error) {
	transfers, err := r.store.GetTransfersWithUser(c, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNoTransfers
		}
		return nil, err
	}

	ans := make([]*db.Transfer, len(transfers))
	for i := range transfers {
		ans[i] = &transfers[i]
	}

	return ans, nil
}
