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
	"github.com/redis/go-redis/v9"
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

	// Kết nối Redis (bắt buộc cho idempotency)
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,     // ví dụ: "localhost:6379"
		Password: cfg.Redis.Password, // nếu có
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal("redis connect failed", zap.Error(err))
	}
	defer rdb.Close()

	// Khởi tạo repositories
	paymentRepo := repository.NewPaymentRepository(db)
	bookingRepo := repository.NewBookingRepository(db) 

	// Khởi tạo PaymentService với config VNPay từ cfg
	svc := service.NewPaymentService(
		paymentRepo,
		bookingRepo,
		rdb,
		cfg.VNPay.PayURL,    // "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"
		cfg.VNPay.TmnCode,   // ví dụ: "YOURTMN00"
		cfg.VNPay.HashSecret, // secret từ VNPay
		cfg.VNPay.ReturnURL, // "http://localhost:8080/api/payments/callback"
	)

	// Handlers (inject repos/services when implemented)
	authHandler := &handler.AuthHandler{}
	cinemaHandler := &handler.CinemaHandler{}
	bookingHandler := &handler.BookingHandler{}
	paymentHandler := handler.NewPaymentHandler(svc)

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

		// Cinema & showtimes (public for listing)
		r.Get("/cinemas", cinemaHandler.ListCinemas)
		r.Get("/cinemas/{id}", cinemaHandler.GetCinema)
		r.Get("/showtimes", cinemaHandler.ListShowtimes)
		r.Get("/showtimes/{id}/seats", bookingHandler.GetTakenSeats)

		// Protected: booking & payment (require JWT)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
			r.Post("/bookings", bookingHandler.CreateBooking)
			r.Get("/bookings/{id}", bookingHandler.GetBooking)
			//r.Post("/payments", paymentHandler.CreatePayment)  Tạm thời không cần token jwt để test api trước
			r.Get("/payments/{id}", paymentHandler.GetPayment)
		})
		r.Post("/payments", paymentHandler.CreatePayment)

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
