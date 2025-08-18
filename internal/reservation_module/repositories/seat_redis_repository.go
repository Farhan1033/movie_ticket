package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SeatRedisRepository interface {
	HoldSeats(ctx context.Context, scheduleID string, userID string, seats []string, ttl time.Duration) error
	ReleaseSeats(ctx context.Context, scheduleID string, seats []string) error
	IsSeatAvailable(ctx context.Context, scheduleID string, seat string) (bool, error)
	ConfirmSeats(ctx context.Context, scheduleID string, seats []string) error
	GetLockedSeats(ctx context.Context, scheduleID string) (map[string]string, error)
}

type seatRedisRepository struct {
	redis *redis.Client
}

func NewSeatRedisRepository(r *redis.Client) SeatRedisRepository {
	return &seatRedisRepository{redis: r}
}

// HoldSeats menggunakan Lua script untuk atomic operation
func (r *seatRedisRepository) HoldSeats(ctx context.Context, scheduleID string, userID string, seats []string, ttl time.Duration) error {
	key := fmt.Sprintf("reservation:%s", scheduleID)

	// Lua script untuk atomic check and set
	luaScript := `
		local key = KEYS[1]
		local ttl = ARGV[1]
		local userID = ARGV[2]
		
		-- Check if any seats are already taken
		for i = 3, #ARGV do
			local seat = ARGV[i]
			if redis.call('HEXISTS', key, seat) == 1 then
				return {err = 'seat ' .. seat .. ' already taken'}
			end
		end
		
		-- Set all seats atomically
		local data = {}
		for i = 3, #ARGV do
			local seat = ARGV[i]
			table.insert(data, seat)
			table.insert(data, userID)
		end
		
		redis.call('HMSET', key, unpack(data))
		redis.call('EXPIRE', key, ttl)
		
		return {ok = 'success'}
	`

	args := make([]interface{}, 0, len(seats)+2)
	args = append(args, int(ttl.Seconds()), userID)
	for _, seat := range seats {
		args = append(args, seat)
	}

	result, err := r.redis.Eval(ctx, luaScript, []string{key}, args...).Result()
	if err != nil {
		return err
	}

	// Check result
	if resultMap, ok := result.(map[interface{}]interface{}); ok {
		if errMsg, exists := resultMap["err"]; exists {
			return fmt.Errorf("%v", errMsg)
		}
	}

	return nil
}

func (r *seatRedisRepository) ReleaseSeats(ctx context.Context, scheduleID string, seats []string) error {
	key := fmt.Sprintf("reservation:%s", scheduleID)

	if len(seats) == 0 {
		return nil
	}

	fields := make([]string, len(seats))
	for i, seat := range seats {
		fields[i] = seat
	}

	return r.redis.HDel(ctx, key, fields...).Err()
}

func (r *seatRedisRepository) IsSeatAvailable(ctx context.Context, scheduleID string, seat string) (bool, error) {
	key := fmt.Sprintf("reservation:%s", scheduleID)

	exists, err := r.redis.HExists(ctx, key, seat).Result()
	if err != nil {
		return false, err
	}

	return !exists, nil
}

func (r *seatRedisRepository) ConfirmSeats(ctx context.Context, scheduleID string, seats []string) error {
	tempKey := fmt.Sprintf("reservation:%s", scheduleID)
	confirmKey := fmt.Sprintf("confirmed:%s", scheduleID)

	luaScript := `
		local tempKey = KEYS[1]
		local confirmKey = KEYS[2]

		local data = {}
		local count = 0
		for i = 1, #ARGV do
			local seat = ARGV[i]
			local userID = redis.call('HGET', tempKey, seat)
			if userID then
				table.insert(data, seat)
				table.insert(data, userID)
				redis.call('HDEL', tempKey, seat)
				count = count + 1
			end
		end

		if #data > 0 then
			redis.call('HMSET', confirmKey, unpack(data))
		end

		-- kalau sudah tidak ada field, hapus key reservation:<schedule_id>
		if redis.call('HLEN', tempKey) == 0 then
			redis.call('DEL', tempKey)
		end

		return count
	`

	args := make([]interface{}, len(seats))
	for i, seat := range seats {
		args[i] = seat
	}

	_, err := r.redis.Eval(ctx, luaScript, []string{tempKey, confirmKey}, args...).Result()
	return err
}

func (r *seatRedisRepository) GetLockedSeats(ctx context.Context, scheduleID string) (map[string]string, error) {
	key := fmt.Sprintf("reservation:%s", scheduleID)

	result, err := r.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return result, nil
}
