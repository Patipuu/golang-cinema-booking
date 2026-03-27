package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"booking_cinema_golang/internal/config"
	"booking_cinema_golang/internal/database"
	"booking_cinema_golang/internal/handler"
	"booking_cinema_golang/internal/middleware"
	"booking_cinema_golang/internal/repository"
	"booking_cinema_golang/internal/service"
	"booking_cinema_golang/internal/utils"

	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load: %v\n", err)
		os.Exit(1)
	}

	logger, err := utils.NewLogger(cfg.Server.Env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.Database.DSN())
	if err != nil {
		logger.Fatal("database connect", zap.Error(err))
	}
	defer db.Close()

	// Wire dependencies
	userRepo := repository.NewUserRepository(db.Pool)
	emailSvc := service.NewEmailService(
		cfg.SMTP.Host, cfg.SMTP.Port,
		cfg.SMTP.User, cfg.SMTP.Password, cfg.SMTP.From,
	)
	authSvc := service.NewAuthService(
		userRepo, emailSvc,
		cfg.JWT.Secret, cfg.JWT.ExpiryHours, cfg.OTP.ExpiryMinutes,
	)
	authHandler := handler.NewAuthHandler(authSvc)
	cinemaHandler := &handler.CinemaHandler{}
	bookingRepo := repository.NewBookingRepository(db)
	bookingSvc := service.NewBookingService(bookingRepo)
	bookingHandler := handler.NewBookingHandler(bookingSvc)
	paymentHandler := &handler.PaymentHandler{}
	userSvc := service.NewUserService(userRepo, cfg)
	userHandler := handler.NewUserHandler(userSvc)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"hello"}`))
		})

		// Auth (public)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/verify-otp", authHandler.VerifyOTP)
		r.Post("/auth/resend-verification", authHandler.ResendVerification)

		// Cinema & showtimes (public for listing)
		r.Get("/cinemas", cinemaHandler.ListCinemas)
		r.Get("/cinemas/{id}", cinemaHandler.GetCinema)
		r.Get("/showtimes", cinemaHandler.ListShowtimes)
		r.Get("/showtimes/{id}/seats", bookingHandler.GetTakenSeats)

		// Protected: booking & payment (require JWT)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWT.Secret))

			r.Get("/users/me", userHandler.GetProfile)
			r.Put("/users/me", userHandler.UpdateProfile)
			r.Put("/users/me/password", userHandler.ChangePassword)

			r.Post("/bookings", bookingHandler.CreateBooking)
			r.Get("/bookings/{id}", bookingHandler.GetBooking)
			r.Post("/payments", paymentHandler.CreatePayment)
			r.Get("/payments/{id}", paymentHandler.GetPayment)
		})
	})

	addr := ":" + cfg.Server.Port
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		logger.Info("server started", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", zap.Error(err))
	}
	logger.Info("server stopped")
}
