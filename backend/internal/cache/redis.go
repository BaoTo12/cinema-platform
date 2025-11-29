package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var ctx = context.Background()

func Connect() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	Client = redis.NewClient(opt)

	// Test connection
	_, err = Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Redis connected successfully")
	return nil
}

// LockSeat locks a seat for a specific showtime and session
func LockSeat(showtimeID, seatID, sessionToken string, expiry time.Duration) (bool, error) {
	key := fmt.Sprintf("lock:showtime:%s:seat:%s", showtimeID, seatID)
	
	// Set only if key doesn't exist (NX flag)
	result, err := Client.SetNX(ctx, key, sessionToken, expiry).Result()
	if err != nil {
		return false, err
	}
	
	return result, nil
}

// UnlockSeat releases a seat lock if it belongs to the given session
func UnlockSeat(showtimeID, seatID, sessionToken string) error {
	key := fmt.Sprintf("lock:showtime:%s:seat:%s", showtimeID, seatID)
	
	// Lua script to ensure we only delete if the value matches
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	
	_, err := Client.Eval(ctx, script, []string{key}, sessionToken).Result()
	return err
}

// GetSeatLock checks if a seat is locked and returns the session token
func GetSeatLock(showtimeID, seatID string) (string, error) {
	key := fmt.Sprintf("lock:showtime:%s:seat:%s", showtimeID, seatID)
	return Client.Get(ctx, key).Result()
}

// LockMultipleSeats attempts to lock multiple seats atomically
func LockMultipleSeats(showtimeID string, seatIDs []string, sessionToken string, expiry time.Duration) (bool, error) {
	pipe := Client.Pipeline()
	
	for _, seatID := range seatIDs {
		key := fmt.Sprintf("lock:showtime:%s:seat:%s", showtimeID, seatID)
		pipe.SetNX(ctx, key, sessionToken, expiry)
	}
	
	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	
	// Check if all locks were acquired
	for _, result := range results {
		if result.(*redis.BoolCmd).Val() == false {
			// Rollback: release any acquired locks
			for _, seatID := range seatIDs {
				UnlockSeat(showtimeID, seatID, sessionToken)
			}
			return false, nil
		}
	}
	
	return true, nil
}

// UnlockMultipleSeats releases multiple seat locks
func UnlockMultipleSeats(showtimeID string, seatIDs []string, sessionToken string) error {
	for _, seatID := range seatIDs {
		if err := UnlockSeat(showtimeID, seatID, sessionToken); err != nil {
			return err
		}
	}
	return nil
}

// SetWithExpiry sets a key with expiration
func SetWithExpiry(key string, value interface{}, expiry time.Duration) error {
	return Client.Set(ctx, key, value, expiry).Err()
}

// Get retrieves a value by key
func Get(key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

// Delete removes a key
func Delete(key string) error {
	return Client.Del(ctx, key).Err()
}

func Close() error {
	return Client.Close()
}
