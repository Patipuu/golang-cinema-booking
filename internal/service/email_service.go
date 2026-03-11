package service

// EmailService defines email-related operations.
type EmailService interface {
	SendVerificationEmail(to string) error
	SendBookingConfirmation(to string) error
}

