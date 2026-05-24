-- name: CreateUserRoom :one
-- 创建用户-寝室关联
INSERT INTO user_rooms (
	username, room_id, threshold, is_enabled
) VALUES (
	$1, $2, $3, $4
)
RETURNING *;

-- name: GetUserRoom :one
-- 查询单个关联
SELECT * FROM user_rooms
WHERE username = $1
	AND room_id = $2
LIMIT 1;

-- name: ListUserRoomsByUser :many
-- 查询某个用户的全部关联
SELECT * FROM user_rooms
WHERE username = $1
ORDER BY room_id ASC
LIMIT $2
OFFSET $3;

-- name: ListUserRoomsByRoom :many
-- 查询某个寝室的全部关联
SELECT * FROM user_rooms
WHERE room_id = $1
ORDER BY username ASC;

-- name: ListUserRooms :many
-- 分页查询关联
SELECT * FROM user_rooms
ORDER BY username ASC, room_id ASC
LIMIT $1
OFFSET $2;

-- name: UpdateUserRoom :one
-- 更新关联阈值和开关
UPDATE user_rooms
SET threshold = coalesce(sqlc.narg(threshold), threshold),
	is_enabled = coalesce(sqlc.narg(is_enabled), is_enabled)
WHERE username = $1
	AND room_id = $2
RETURNING *;

-- name: UpdateUserRoomLastNotifiedAt :one
-- 更新最后通知时间
UPDATE user_rooms
SET last_notified_at = $3
WHERE username = $1
	AND room_id = $2
RETURNING *;

-- name: DeleteUserRoom :exec
-- 删除关联
DELETE FROM user_rooms
WHERE username = $1
	AND room_id = $2;

-- name: CountUserRooms :one
-- 统计关联总数
SELECT COUNT(*) FROM user_rooms;

-- name: ListDueUserRooms :many
-- 查询需要发送通知的关联（开启且上次通知时间超过 24 小时）
SELECT username, room_id, threshold 
FROM user_rooms 
WHERE is_enabled = true 
  AND last_notified_at < (now() - interval '24 hours');