package postgresrepo

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/myacey/avito-shop/db/sqlc"
	"github.com/myacey/avito-shop/internal/repository"
)

type PostgresUserRepo struct {
	store db.Querier
}

func NewPostgresUserRepo(store db.Querier) repository.UserRepository {
	return &PostgresUserRepo{
		store: store,
	}
}

func (r *PostgresUserRepo) CreateUser(c context.Context, username, password string) (*db.User, error) {
	arg := db.CreateUserParams{
		Username: username,
		Password: password,
	}
	usr, err := r.store.CreateUser(c, arg)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, repository.ErrUserAlreadyExists
		}
		return nil, err
	}

	return &usr, nil
}

func (r *PostgresUserRepo) GetUser(c context.Context, username string) (*db.User, error) {
	usr, err := r.store.GetUser(c, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return &usr, nil
}

// Should be called only in transactions.
func (r *PostgresUserRepo) GetUserForUpdate(c context.Context, username string) (*db.User, error) {
	usr, err := r.store.GetUserForUpdate(c, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return &usr, nil
}

func (r *PostgresUserRepo) UpdateBalance(c context.Context, userID int32, newCointCount int32) (*db.User, error) {
	arg := db.UpdateUserBalanceParams{
		UserID: userID,
		Coins:  newCointCount,
	}

	usr, err := r.store.UpdateUserBalance(c, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return &usr, nil
}

func (r *PostgresUserRepo) UpdateTwoUsersBalance(c context.Context, fromUsername, toUsername string, coinsAmount int32) ([]*db.User, error) {
	usrs, err := r.store.UpdateTwoUsersBalance(c, db.UpdateTwoUsersBalanceParams{
		FromUsername: fromUsername,
		ToUsername:   toUsername,
		Coins:        coinsAmount,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	ans := make([]*db.User, len(usrs))
	for i := range usrs {
		ans[i] = &usrs[i]
	}

	return ans, nil
}
