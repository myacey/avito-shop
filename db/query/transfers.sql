-- name: CreateMoneyTransfer :one
INSERT INTO Transfers (from_username, to_username, amount)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetTransfersWithUser :many
SELECT * FROM Transfers
WHERE from_username=sqlc.arg(username) OR to_username=sqlc.arg(username)
FOR SHARE;