-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: PostgreSQL
-- Generated at: 2026-03-21T08:27:37.446Z

CREATE TABLE "rooms" (
  "id" BIGSERIAL PRIMARY KEY,
  "name" varchar NOT NULL,
  "area_id" varchar NOT NULL,
  "building_code" varchar NOT NULL,
  "floor_code" varchar NOT NULL,
  "room_code" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "electricity_records" (
  "id" BIGSERIAL PRIMARY KEY,
  "room_id" bigint NOT NULL,
  "balance" bigint NOT NULL,
  "recorded_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "users" (
  "username" varchar PRIMARY KEY NOT NULL,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "user_room_notifications" (
  "username" varchar NOT NULL,
  "room_id" bigint NOT NULL,
  "threshold" int NOT NULL DEFAULT 10,
  "is_enabled" boolean NOT NULL DEFAULT true,
  "last_notified_at" timestamptz,
  PRIMARY KEY ("username", "room_id")
);

CREATE UNIQUE INDEX ON "electricity_records" ("room_id", "recorded_at");

CREATE INDEX ON "electricity_records" ("recorded_at");

COMMENT ON TABLE "rooms" IS '存储需要监控的寝室配置信息';

COMMENT ON TABLE "electricity_records" IS '存储抓取的电费流水数据';

COMMENT ON COLUMN "user_room_notifications"."threshold" IS '预警阈值，单位: 元';

COMMENT ON COLUMN "user_room_notifications"."last_notified_at" IS '上次发送邮件的时间';

ALTER TABLE "electricity_records" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_room_notifications" ADD FOREIGN KEY ("username") REFERENCES "users" ("username") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_room_notifications" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;
