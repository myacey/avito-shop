-- name: CreateUser :one
INSERT INTO Users (
    username,
    password
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetUser :one
SELECT * FROM Users
WHERE username = $1
LIMIT 1 FOR SHARE;

-- name: GetUserForUpdate :one
SELECT * FROM Users
WHERE username = $1
LIMIT 1
FOR UPDATE;

-- name: GetUserViaID :one
SELECT * FROM Users
WHERE user_id = $1
LIMIT 1 FOR SHARE;

-- name: UpdateTwoUsersBalance :many
UPDATE Users
SET coins = CASE 
    WHEN username = sqlc.arg(from_username) THEN coins - $1
    WHEN username = sqlc.arg(to_username) THEN coins + $1
END
WHERE USERNAME IN (sqlc.arg(from_username), sqlc.arg(to_username))
RETURNING *;


-- name: UpdateUserBalance :one
UPDATE Users
SET coins = $2
WHERE user_id = $1
RETURNING *;
