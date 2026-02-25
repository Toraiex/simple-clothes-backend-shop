ALTER TABLE users
ADD COLUMN last_verification_otp_sent_at TIMESTAMPTZ,
ADD COLUMN verification_otp_resend_count INT DEFAULT 0,

ADD COLUMN last_reset_otp_sent_at TIMESTAMPTZ,
ADD COLUMN reset_otp_resend_count INT DEFAULT 0;