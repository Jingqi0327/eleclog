-- name: CreateElectricityRecord :one
-- 插入一条电费记录
INSERT INTO electricity_records (
    room_id, balance
) VALUES (
    $1, $2
)
RETURNING *;

-- name: CreateElectricityRecordTest :one
-- 用于插入测试数据的电费记录
INSERT INTO electricity_records (
    room_id, balance, recorded_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetRecordsByHourRange :many
-- 获取指定时间范围内的每小时记录
SELECT * FROM electricity_records
WHERE room_id = $1 
  AND recorded_at BETWEEN $2 AND $3
ORDER BY recorded_at ASC;


-- name: GetLatestBalance :one
-- 获取最新的余额记录
SELECT * FROM electricity_records
WHERE room_id = $1
ORDER BY recorded_at DESC
LIMIT 1;