DB_URL=postgresql://root:123456@localhost:5432/elec_log?sslmode=disable
TEST_DB_URL=postgresql://root:123456@localhost:5432/test_elec_log?sslmode=disable

createdb:
	docker exec -it postgres3.23 createdb --username=root --owner=root elec_log

dropdb:
	docker exec -it postgres3.23 dropdb elec_log

createtestdb:
	docker exec -it postgres3.23 createdb --username=root --owner=root test_elec_log

droptestdb:
	docker exec -it postgres3.23 dropdb test_elec_log

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratetestup:
	migrate -path db/migration -database "$(TEST_DB_URL)" -verbose up

migratetestdown:
	migrate -path db/migration -database "$(TEST_DB_URL)" -verbose down

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

server:
	go run main.go