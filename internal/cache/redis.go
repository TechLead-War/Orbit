package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache handles caching using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisCache{
		client: client,
	}
}

// Get retrieves a value from the cache
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Set stores a value in the cache with an expiration time
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, expiration).Err()
}

// Del removes a value from the cache
func (c *RedisCache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// FlushAll removes all keys from the cache
func (c *RedisCache) FlushAll(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Ping checks if Redis is available
func (c *RedisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// GetMulti retrieves multiple values from the cache
func (c *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	pipeline := c.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd, len(keys))

	for _, key := range keys {
		cmds[key] = pipeline.Get(ctx, key)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	result := make(map[string]interface{}, len(keys))
	for key, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, err
		}

		var value interface{}
		if err := json.Unmarshal([]byte(val), &value); err != nil {
			return nil, err
		}
		result[key] = value
	}

	return result, nil
}

// SetMulti stores multiple values in the cache
func (c *RedisCache) SetMulti(ctx context.Context, values map[string]interface{}, expiration time.Duration) error {
	pipeline := c.client.Pipeline()

	for key, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		pipeline.Set(ctx, key, data, expiration)
	}

	_, err := pipeline.Exec(ctx)
	return err
}

// DelMulti removes multiple values from the cache
func (c *RedisCache) DelMulti(ctx context.Context, keys []string) error {
	return c.client.Del(ctx, keys...).Err()
}

// SetNX sets a value if the key doesn't exist
func (c *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	return c.client.SetNX(ctx, key, data, expiration).Result()
}

// Incr increments a counter
func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrBy increments a counter by a specific value
func (c *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decr decrements a counter
func (c *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrBy decrements a counter by a specific value
func (c *RedisCache) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// Expire sets an expiration time on a key
func (c *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(ctx, key, expiration).Result()
}

// TTL gets the remaining time to live of a key
func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Keys gets all keys matching a pattern
func (c *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// Type gets the type of a key
func (c *RedisCache) Type(ctx context.Context, key string) (string, error) {
	return c.client.Type(ctx, key).Result()
}

// Exists checks if a key exists
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// ExistsMulti checks if multiple keys exist
func (c *RedisCache) ExistsMulti(ctx context.Context, keys []string) (map[string]bool, error) {
	pipeline := c.client.Pipeline()
	cmds := make(map[string]*redis.IntCmd, len(keys))

	for _, key := range keys {
		cmds[key] = pipeline.Exists(ctx, key)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(keys))
	for key, cmd := range cmds {
		n, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		result[key] = n > 0
	}

	return result, nil
}
