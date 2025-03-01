// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: transfers.sql

package db

import (
	"context"
)

const createMoneyTransfer = `-- name: CreateMoneyTransfer :one
INSERT INTO Transfers (from_username, to_username, amount)
VALUES ($1, $2, $3)
RETURNING transfer_id, from_username, to_username, amount
`

type CreateMoneyTransferParams struct {
	FromUsername string `json:"from_username"`
	ToUsername   string `json:"to_username"`
	Amount       int32  `json:"amount"`
}

func (q *Queries) CreateMoneyTransfer(ctx context.Context, arg CreateMoneyTransferParams) (Transfer, error) {
	row := q.db.QueryRowContext(ctx, createMoneyTransfer, arg.FromUsername, arg.ToUsername, arg.Amount)
	var i Transfer
	err := row.Scan(
		&i.TransferID,
		&i.FromUsername,
		&i.ToUsername,
		&i.Amount,
	)
	return i, err
}

const getTransfersWithUser = `-- name: GetTransfersWithUser :many
SELECT transfer_id, from_username, to_username, amount FROM Transfers
WHERE from_username=$1 OR to_username=$1
FOR SHARE
`

func (q *Queries) GetTransfersWithUser(ctx context.Context, username string) ([]Transfer, error) {
	rows, err := q.db.QueryContext(ctx, getTransfersWithUser, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Transfer{}
	for rows.Next() {
		var i Transfer
		if err := rows.Scan(
			&i.TransferID,
			&i.FromUsername,
			&i.ToUsername,
			&i.Amount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
