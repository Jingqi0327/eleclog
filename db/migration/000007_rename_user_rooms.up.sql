ALTER TABLE "user_room_notifications" RENAME TO "user_rooms";
ALTER TABLE "user_rooms" ALTER COLUMN "is_enabled" SET DEFAULT false;
