package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"order-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type OrderCache interface {
	Get(ctx context.Context, orderID string) (domain.Order, bool)
	Set(ctx context.Context, order domain.Order)
	Delete(ctx context.Context, orderID string)
}

type redisOrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisOrderCache() OrderCache {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	ttlSecs := 300
	if v := os.Getenv("CACHE_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			ttlSecs = n
		}
	}

	client := redis.NewClient(&redis.Options{Addr: addr})
	log.Printf("[Cache] Connected to Redis at %s, TTL=%ds", addr, ttlSecs)

	return &redisOrderCache{
		client: client,
		ttl:    time.Duration(ttlSecs) * time.Second,
	}
}

func cacheKey(orderID string) string {
	return fmt.Sprintf("order:%s", orderID)
}

func (c *redisOrderCache) Get(ctx context.Context, orderID string) (domain.Order, bool) {
	val, err := c.client.Get(ctx, cacheKey(orderID)).Result()
	if err != nil {
		return domain.Order{}, false
	}

	var order domain.Order
	if err := json.Unmarshal([]byte(val), &order); err != nil {
		log.Printf("[Cache] Failed to unmarshal order %s: %v", orderID, err)
		return domain.Order{}, false
	}

	log.Printf("[Cache] HIT for order %s", orderID)
	return order, true
}

func (c *redisOrderCache) Set(ctx context.Context, order domain.Order) {
	data, err := json.Marshal(order)
	if err != nil {
		log.Printf("[Cache] Failed to marshal order %s: %v", order.ID, err)
		return
	}

	if err := c.client.Set(ctx, cacheKey(order.ID), data, c.ttl).Err(); err != nil {
		log.Printf("[Cache] Failed to set order %s: %v", order.ID, err)
		return
	}

	log.Printf("[Cache] SET order %s (TTL=%s)", order.ID, c.ttl)
}

func (c *redisOrderCache) Delete(ctx context.Context, orderID string) {
	if err := c.client.Del(ctx, cacheKey(orderID)).Err(); err != nil {
		log.Printf("[Cache] Failed to delete order %s from cache: %v", orderID, err)
		return
	}
	log.Printf("[Cache] INVALIDATED order %s", orderID)
}
