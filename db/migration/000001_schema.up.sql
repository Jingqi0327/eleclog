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

CREATE UNIQUE INDEX ON "electricity_records" ("room_id", "recorded_at");

CREATE INDEX ON "electricity_records" ("recorded_at");

COMMENT ON TABLE "rooms" IS '存储需要监控的寝室配置信息';

COMMENT ON TABLE "electricity_records" IS '存储抓取的电费流水数据';

ALTER TABLE "electricity_records" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id") DEFERRABLE INITIALLY IMMEDIATE;
