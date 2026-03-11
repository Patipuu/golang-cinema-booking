package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateOTP returns a 6-digit numeric OTP string.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// OTPExpiry returns expiry time for OTP (e.g. 10 minutes from now).
func OTPExpiry(validMinutes int) time.Time {
	return time.Now().Add(time.Duration(validMinutes) * time.Minute)
}

// IsOTPExpired checks if expiry time has passed.
func IsOTPExpired(expiry *time.Time) bool {
	if expiry == nil {
		return true
	}
	return time.Now().After(*expiry)
}
