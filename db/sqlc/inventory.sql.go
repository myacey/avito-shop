// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: inventory.sql

package db

import (
	"context"
)

const buyItem = `-- name: BuyItem :exec
INSERT INTO Inventory (user_id, item_type)
VALUES ($1, $2)
ON CONFLICT (user_id, item_type)
DO UPDATE SET quantity = Inventory.quantity + 1
`

type BuyItemParams struct {
	UserID   int32  `json:"user_id"`
	ItemType string `json:"item_type"`
}

func (q *Queries) BuyItem(ctx context.Context, arg BuyItemParams) error {
	_, err := q.db.ExecContext(ctx, buyItem, arg.UserID, arg.ItemType)
	return err
}

const getInventory = `-- name: GetInventory :many
SELECT inventory_id, user_id, item_type, quantity FROM Inventory
WHERE user_id=$1
FOR SHARE
`

func (q *Queries) GetInventory(ctx context.Context, userID int32) ([]Inventory, error) {
	rows, err := q.db.QueryContext(ctx, getInventory, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Inventory{}
	for rows.Next() {
		var i Inventory
		if err := rows.Scan(
			&i.InventoryID,
			&i.UserID,
			&i.ItemType,
			&i.Quantity,
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
