CREATE TABLE "user_room_notifications" (
	"username" varchar NOT NULL,
	"room_id" bigint NOT NULL,
	"threshold" int NOT NULL DEFAULT 10,
	"is_enabled" boolean NOT NULL DEFAULT true,
	"last_notified_at" timestamptz NOT NULL DEFAULT '0001-01-01',
	PRIMARY KEY ("username", "room_id")
);

COMMENT ON COLUMN "user_room_notifications"."threshold" IS '预警阈值，单位: 元';

COMMENT ON COLUMN "user_room_notifications"."last_notified_at" IS '上次发送邮件的时间';

ALTER TABLE "user_room_notifications"
ADD CONSTRAINT "user_room_notifications_username_fkey"
FOREIGN KEY ("username")
REFERENCES "users" ("username")
ON DELETE CASCADE
DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_room_notifications"
ADD CONSTRAINT "user_room_notifications_room_id_fkey"
FOREIGN KEY ("room_id")
REFERENCES "rooms" ("id")
ON DELETE CASCADE
DEFERRABLE INITIALLY IMMEDIATE;
