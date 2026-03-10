-- เพิ่มคอลัมน์ email (ห้ามซ้ำ) และ is_verified (ค่าเริ่มต้นเป็น false)
ALTER TABLE users 
ADD COLUMN email TEXT UNIQUE,
ADD COLUMN is_verified BOOLEAN DEFAULT false;

ALTER TABLE users 
ADD COLUMN otp_code VARCHAR(6),
ADD COLUMN otp_expires_at TIMESTAMP WITH TIME ZONE;