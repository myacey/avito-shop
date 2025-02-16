-- name: GetItemFromStore :one
SELECT * FROM Items
WHERE item_type = $1
LIMIT 1 
FOR SHARE;