--创建用户
-- name: CreateUser :one
INSERT INTO users (
  username, hashed_password, full_name, role, email
) VALUES (
  $1, $2, $3, $4,$5
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
  role = coalesce(sqlc.narg(role), role),
  email = coalesce(sqlc.narg(email), email)
WHERE 
  username = $1
RETURNING *;

-- name: CountUsers :one
-- 查询用户个数
SELECT COUNT(*) FROM users;


-- name: ListUsers :many
-- 分页查询用户
SELECT * FROM users
LIMIT $1 
OFFSET $2;


-- name: DeleteUser :exec
-- 删除用户
DELETE FROM users
WHERE username = $1;