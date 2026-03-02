package domain

type EmailService interface {
	SendVerificationEmail(toEmail string, otpCode string) error
	SendPasswordResetEmail(toEmail string, otpCode string) error
}
