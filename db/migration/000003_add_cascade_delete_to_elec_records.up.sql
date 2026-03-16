ALTER TABLE electricity_records
DROP CONSTRAINT IF EXISTS electricity_records_room_id_fkey;

ALTER TABLE electricity_records
ADD CONSTRAINT electricity_records_room_id_fkey
FOREIGN KEY (room_id)
REFERENCES rooms(id)
ON DELETE CASCADE
DEFERRABLE INITIALLY IMMEDIATE;
