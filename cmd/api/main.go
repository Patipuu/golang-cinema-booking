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
	"booking_cinema_golang/internal/infrastructure/redis"
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

	// 1. Infrastructure
	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		logger.Fatal("redis connect", zap.Error(err))
	}
	defer rdb.Close()

	hub := handler.NewHub()
	go hub.Run()

	// 2. Repositories
	userRepo := repository.NewUserRepository(db.Pool)
	catalogRepo := repository.NewCatalogRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	paymentMethodRepo := repository.NewPaymentMethodRepository(db)

	// 3. Services
	emailSvc := service.NewEmailService(
		cfg.SMTP.Host, cfg.SMTP.Port,
		cfg.SMTP.User, cfg.SMTP.Password, cfg.SMTP.From,
	)
	authSvc := service.NewAuthService(
		userRepo, emailSvc,
		cfg.JWT.Secret, cfg.JWT.ExpiryHours, cfg.OTP.ExpiryMinutes,
	)
	catalogSvc := service.NewCatalogService(catalogRepo)
	pricingSvc := service.NewPricingService()
	bookingSvc := service.NewBookingService(bookingRepo, rdb, pricingSvc, hub)
	paymentSvc := service.NewPaymentService(
		paymentRepo, paymentMethodRepo, bookingRepo, rdb.GetRDB(),
		cfg.VNPay.PayURL, cfg.VNPay.TmnCode, cfg.VNPay.HashSecret, cfg.VNPay.ReturnURL,
	)

	// 4. Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	catalogHandler := handler.NewCatalogHandler(catalogSvc)
	adminHandler := handler.NewAdminHandler(catalogSvc, authSvc, bookingSvc)
	bookingHandler := handler.NewBookingHandler(bookingSvc)
	paymentHandler := handler.NewPaymentHandler(paymentSvc, bookingSvc)

	// 5. Routing
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	isMaintenance := true
	r.Use(middleware.MaintenanceMiddleware(isMaintenance))

	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeWs(hub, w, r)
	})

	// Standardized API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public Auth
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/verify-otp", authHandler.VerifyOTP)
		r.Post("/logout", authHandler.Logout)

		// Movie
		r.Get("/movies", catalogHandler.ListMovies)
		r.Get("/movies/{id}", catalogHandler.GetMovie)

		// Cinema
		r.Get("/cinemas", catalogHandler.ListCinemas)
		r.Get("/rooms", catalogHandler.ListRooms)
		r.Get("/seats/room/{id}", catalogHandler.ListSeats)

		// Showtime
		r.Get("/showtimes", catalogHandler.ListShowtimes)
		r.Get("/seats/showtime/{id}", bookingHandler.GetTakenSeats)

		// Callback cho cổng thanh toán
		r.Get("/payments/callback", paymentHandler.HandleVNPayCallback)
		// Protected - User
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
			r.Post("/bookings", bookingHandler.CreateBooking)
			r.Post("/bookings/lock", bookingHandler.LockSeat)
			r.Post("/bookings/unlock", bookingHandler.UnlockSeat)
			r.Get("/bookings/{id}", bookingHandler.GetBooking)
			r.Get("/bookings/my", bookingHandler.ListMyBookings)
			r.Delete("/bookings/{id}", bookingHandler.CancelBooking)

			r.Post("/payment", paymentHandler.CreatePayment)

		})

		// Protected - Admin
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
			// Admin Movie/Cinema
			r.Get("/cinemas", adminHandler.ListAllCinemas)
			r.Post("/cinemas", adminHandler.CreateCinema)
			r.Put("/cinemas/{id}", adminHandler.UpdateCinema)
			r.Delete("/cinemas/{id}", adminHandler.DeleteCinema)
			r.Get("/rooms", adminHandler.ListRooms)

			r.Get("/movies", adminHandler.ListAllMovies)
			r.Post("/movies", adminHandler.CreateMovie)
			r.Put("/movies/{id}", adminHandler.UpdateMovie)

			r.Get("/showtimes", adminHandler.ListAllShowtimes)
			r.Post("/showtimes", adminHandler.CreateShowtime)
			r.Put("/showtimes/{id}", adminHandler.UpdateShowtime)
			r.Delete("/showtimes/{id}", adminHandler.DeleteShowtime)
			r.Get("/stats", adminHandler.GetStats)

			// Admin Users
			r.Get("/users", adminHandler.ListUsers)
			r.Put("/users/{id}/role", adminHandler.UpdateUserRole)
		})
	})

	// Static Frontend
	fs := http.FileServer(http.Dir("./frontend"))
	r.Handle("/*", fs)

	addr := ":" + cfg.Server.Port
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		logger.Info("Cinema Server Started", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	logger.Info("Cinema Server Stopped")
}
