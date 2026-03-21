-- name: CreateUserRoomNotification :one
-- 创建用户-寝室通知订阅
INSERT INTO user_room_notifications (
	username, room_id, threshold
) VALUES (
	$1, $2, $3
)
RETURNING *;

-- name: GetUserRoomNotification :one
-- 查询单个通知订阅
SELECT * FROM user_room_notifications
WHERE username = $1
	AND room_id = $2
LIMIT 1;

-- name: ListUserRoomNotificationsByUser :many
-- 查询某个用户的全部通知订阅
SELECT * FROM user_room_notifications
WHERE username = $1
ORDER BY room_id ASC;

-- name: ListUserRoomNotificationsByRoom :many
-- 查询某个寝室的全部通知订阅
SELECT * FROM user_room_notifications
WHERE room_id = $1
ORDER BY username ASC;

-- name: ListUserRoomNotifications :many
-- 分页查询通知订阅
SELECT * FROM user_room_notifications
ORDER BY username ASC, room_id ASC
LIMIT $1
OFFSET $2;

-- name: UpdateUserRoomNotification :one
-- 更新通知阈值和开关
UPDATE user_room_notifications
SET threshold = coalesce(sqlc.narg(threshold), threshold),
	is_enabled = coalesce(sqlc.narg(is_enabled), is_enabled)
WHERE username = $1
	AND room_id = $2
RETURNING *;

-- name: UpdateUserRoomNotificationLastNotifiedAt :one
-- 更新最后通知时间
UPDATE user_room_notifications
SET last_notified_at = $3
WHERE username = $1
	AND room_id = $2
RETURNING *;

-- name: DeleteUserRoomNotification :exec
-- 删除通知订阅
DELETE FROM user_room_notifications
WHERE username = $1
	AND room_id = $2;

-- name: CountUserRoomNotifications :one
-- 统计通知订阅总数
SELECT COUNT(*) FROM user_room_notifications;

-- name: ListDueUserRoomNotifications :many
-- 查询需要发送通知的订阅（开启且上次通知时间超过 24 小时）
SELECT username, room_id, threshold 
FROM user_room_notifications 
WHERE is_enabled = true 
  AND last_notified_at < (now() - interval '24 hours');