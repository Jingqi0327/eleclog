-- name: CreateRoom :one
-- 插入一个要查询的寝室信息
INSERT INTO rooms (
  name, area_id, building_code, floor_code, room_code
) VALUES (
 $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetRoom :one
-- 根据ID查询寝室信息
SELECT * FROM rooms
WHERE id = $1 LIMIT 1;

-- name: ListRooms :many
-- 查询所有寝室信息
SELECT * FROM rooms
ORDER BY name
LIMIT $1 
OFFSET $2;


-- name: UpdateRoom :one
-- 更新寝室信息
UPDATE rooms
SET name = coalesce(sqlc.narg(name), name),
    area_id = coalesce(sqlc.narg(area_id), area_id),
    building_code = coalesce(sqlc.narg(building_code), building_code),
    floor_code = coalesce(sqlc.narg(floor_code), floor_code),
    room_code = coalesce(sqlc.narg(room_code), room_code)
WHERE id = $1
RETURNING *;

-- name: DeleteRoom :exec
-- 删除寝室信息
DELETE FROM rooms
WHERE id = $1;