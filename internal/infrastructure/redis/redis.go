package redis

import (
	"context"
	"fmt"
	"time"

	"booking_cinema_golang/internal/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("Redis connection failed: %v. Starting in-memory miniredis for testing...\n", err)
		mr, err := miniredis.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to start miniredis: %w", err)
		}
		rdb = redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) LockSeat(ctx context.Context, showtimeID, seatID string, expiration time.Duration) (bool, error) {
	key := fmt.Sprintf("seat_lock:%s:%s", showtimeID, seatID)
	return c.rdb.SetNX(ctx, key, "locked", expiration).Result()
}

func (c *Client) UnlockSeat(ctx context.Context, showtimeID, seatID string) error {
	key := fmt.Sprintf("seat_lock:%s:%s", showtimeID, seatID)
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) IsSeatLocked(ctx context.Context, showtimeID, seatID string) (bool, error) {
	key := fmt.Sprintf("seat_lock:%s:%s", showtimeID, seatID)
	n, err := c.rdb.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *Client) GetLockedSeats(ctx context.Context, showtimeID string) ([]string, error) {
	prefix := fmt.Sprintf("seat_lock:%s:", showtimeID)
	keys, err := c.rdb.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return nil, err
	}
	seatIDs := make([]string, 0, len(keys))
	for _, k := range keys {
		seatIDs = append(seatIDs, k[len(prefix):])
	}
	return seatIDs, nil
}

func (c *Client) GetRDB() *redis.Client {
	return c.rdb
}
