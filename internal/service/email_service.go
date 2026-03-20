package service

// EmailService defines email-related operations.
type EmailService interface {
	SendVerificationEmail(to, fullName, otpCode string, expiresInMinutes int) error
	SendBookingConfirmation(to string) error
}
