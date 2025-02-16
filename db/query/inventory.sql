-- name: BuyItem :exec
INSERT INTO Inventory (user_id, item_type)
VALUES ($1, $2)
ON CONFLICT (user_id, item_type)
DO UPDATE SET quantity = Inventory.quantity + 1;

-- name: GetInventory :many
SELECT * FROM Inventory
WHERE user_id=$1
FOR SHARE;
