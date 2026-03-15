--创建用户
-- name: CreateUser :one
INSERT INTO users (
  username, hashed_password, full_name, email
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

--查询单个用户
-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 
LIMIT 1;


--更新用户信息
-- name: UpdateUser :one
UPDATE users
SET 
  hashed_password = coalesce(sqlc.narg(hashed_password), hashed_password),
  full_name = coalesce(sqlc.narg(full_name), full_name),
  email = coalesce(sqlc.narg(email), email)
WHERE 
  username = $1
RETURNING *;