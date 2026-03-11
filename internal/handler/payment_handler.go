package handler

import "net/http"

// PaymentHandler groups HTTP handlers for payments.
type PaymentHandler struct{}

// CreatePayment handles payment creation.
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	// TODO: implement create payment logic.
}

// GetPayment handles fetching a payment.
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	// TODO: implement get payment logic.
}

