ALTER TABLE users
DROP COLUMN IF EXISTS last_verification_otp_sent_at,
DROP COLUMN IF EXISTS verification_otp_resend_count,
DROP COLUMN IF EXISTS last_reset_otp_sent_at,
DROP COLUMN IF EXISTS reset_otp_resend_count;