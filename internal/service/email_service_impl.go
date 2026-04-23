package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

type smtpEmailService struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

type consoleEmailService struct {
	from string
}

// NewEmailService creates an EmailService.
// If SMTP credentials are not configured, it falls back to console logging for local OTP testing.
func NewEmailService(host string, port int, user, password, from string) EmailService {
	if host == "" || user == "" || password == "" {
		return &consoleEmailService{from: from}
	}
	return &smtpEmailService{host: host, port: port, user: user, password: password, from: from}
}

func (s *consoleEmailService) SendVerificationEmail(to, fullName, otpCode string, expiresInMinutes int) error {
	fmt.Printf("[OTP EMAIL] To=%s From=%s Subject=Verify Your Email — Cinema Booking\n", to, s.from)
	fmt.Printf("Hello %s,\nYour verification code is: %s\nIt expires in %d minutes.\n\n", fullName, otpCode, expiresInMinutes)
	return nil
}

func (s *consoleEmailService) SendBookingConfirmation(to string) error {
	fmt.Printf("[BOOKING EMAIL] To=%s From=%s\n", to, s.from)
	return nil
}

func (s *smtpEmailService) SendVerificationEmail(to, fullName, otpCode string, expiresInMinutes int) error {
	subject := "Verify Your Email — Cinema Booking"
	body := fmt.Sprintf(
		"Hello %s,\n\nYour verification code is:\n\n    %s\n\nThis code expires in %d minutes.\n\nIf you did not create an account, please ignore this email.\n\n— Cinema Booking Team",
		fullName, otpCode, expiresInMinutes,
	)
	return s.send(to, subject, body)
}

func (s *smtpEmailService) SendBookingConfirmation(to string) error {
	// TODO: implement booking confirmation email
	return nil
}

// send builds a plain-text email message and delivers it via SMTP.
func (s *smtpEmailService) send(to, subject, body string) error {
	msg := buildMessage(s.from, to, subject, body)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	tlsCfg := &tls.Config{ServerName: s.host}

	if s.port == 465 {
		return s.sendTLS(addr, tlsCfg, to, msg)
	}
	return s.sendSTARTTLS(addr, tlsCfg, to, msg)
}

// sendTLS dials an SSL/TLS connection directly (port 465).
func (s *smtpEmailService) sendTLS(addr string, tlsCfg *tls.Config, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("email: tls dial: %w", err)
	}

	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("email: smtp client: %w", err)
	}
	defer c.Quit()

	// Use custom plain auth that bypasses the TLS check — we already have TLS.
	if err := c.Auth(&tlsPlainAuth{username: s.user, password: s.password}); err != nil {
		return fmt.Errorf("email: auth: %w", err)
	}
	return s.writeMessage(c, to, msg)
}

// sendSTARTTLS connects plain then upgrades to TLS (port 587).
func (s *smtpEmailService) sendSTARTTLS(addr string, tlsCfg *tls.Config, to string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("email: dial: %w", err)
	}
	defer c.Quit()

	if err := c.StartTLS(tlsCfg); err != nil {
		return fmt.Errorf("email: starttls: %w", err)
	}
	if err := c.Auth(smtp.PlainAuth("", s.user, s.password, s.host)); err != nil {
		return fmt.Errorf("email: auth: %w", err)
	}
	return s.writeMessage(c, to, msg)
}

func (s *smtpEmailService) writeMessage(c *smtp.Client, to string, msg []byte) error {
	if err := c.Mail(s.from); err != nil {
		return fmt.Errorf("email: MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("email: RCPT TO: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("email: DATA: %w", err)
	}
	defer w.Close()

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("email: write: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	var sb strings.Builder
	sb.WriteString("From: " + from + "\r\n")
	sb.WriteString("To: " + to + "\r\n")
	sb.WriteString("Subject: " + subject + "\r\n")
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body + "\r\n")
	return []byte(sb.String())
}

// tlsPlainAuth implements smtp.Auth for connections already using TLS (port 465).
// Go's stdlib smtp.PlainAuth refuses to send credentials unless it sees server.TLS=true,
// but smtp.NewClient() doesn't mark c.tls=true even on a TLS conn. This bypasses that check.
type tlsPlainAuth struct {
	username, password string
}

func (a *tlsPlainAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return "PLAIN", []byte("\x00" + a.username + "\x00" + a.password), nil
}

func (a *tlsPlainAuth) Next(_ []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}
