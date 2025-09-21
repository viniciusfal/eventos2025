package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Config contém as configurações do Redis
type Config struct {
	Host            string
	Port            int
	Password        string
	DB              int
	MaxRetries      int
	PoolSize        int
	MinIdleConns    int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ConnMaxLifetime time.Duration
}

// Client representa um cliente Redis
type Client struct {
	client *redis.Client
	logger *zap.Logger
}

// NewClient cria uma nova instância do cliente Redis
func NewClient(cfg Config, logger *zap.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      cfg.MaxRetries,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	})

	// Testar conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis client connected successfully",
		zap.String("addr", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
		zap.Int("db", cfg.DB),
	)

	return &Client{
		client: rdb,
		logger: logger,
	}, nil
}

// Close fecha a conexão com o Redis
func (c *Client) Close() error {
	return c.client.Close()
}

// Ping testa a conexão com o Redis
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Set armazena um valor no cache com TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal value for cache",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		c.logger.Error("Failed to set cache value",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	c.logger.Debug("Cache value set successfully",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
	)
	return nil
}

// Get recupera um valor do cache
func (c *Client) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		c.logger.Error("Failed to get cache value",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get cache value: %w", err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		c.logger.Error("Failed to unmarshal cache value",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	c.logger.Debug("Cache value retrieved successfully", zap.String("key", key))
	return nil
}

// Delete remove um valor do cache
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	err := c.client.Del(ctx, keys...).Err()
	if err != nil {
		c.logger.Error("Failed to delete cache keys",
			zap.Strings("keys", keys),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete cache keys: %w", err)
	}

	c.logger.Debug("Cache keys deleted successfully", zap.Strings("keys", keys))
	return nil
}

// Exists verifica se uma chave existe no cache
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		c.logger.Error("Failed to check cache key existence",
			zap.Strings("keys", keys),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to check key existence: %w", err)
	}

	return count, nil
}

// SetNX armazena um valor apenas se a chave não existir
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("Failed to marshal value for cache",
			zap.String("key", key),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	success, err := c.client.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		c.logger.Error("Failed to set cache value with NX",
			zap.String("key", key),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to set cache value with NX: %w", err)
	}

	if success {
		c.logger.Debug("Cache value set with NX successfully",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
		)
	}

	return success, nil
}

// Increment incrementa um valor numérico no cache
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	value, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to increment cache value",
			zap.String("key", key),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to increment cache value: %w", err)
	}

	c.logger.Debug("Cache value incremented successfully",
		zap.String("key", key),
		zap.Int64("new_value", value),
	)
	return value, nil
}

// IncrementBy incrementa um valor numérico por uma quantidade específica
func (c *Client) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	newValue, err := c.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		c.logger.Error("Failed to increment cache value by amount",
			zap.String("key", key),
			zap.Int64("increment", value),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to increment cache value: %w", err)
	}

	c.logger.Debug("Cache value incremented by amount successfully",
		zap.String("key", key),
		zap.Int64("increment", value),
		zap.Int64("new_value", newValue),
	)
	return newValue, nil
}

// Expire define um TTL para uma chave existente
func (c *Client) Expire(ctx context.Context, key string, ttl time.Duration) error {
	success, err := c.client.Expire(ctx, key, ttl).Result()
	if err != nil {
		c.logger.Error("Failed to set cache key expiration",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return fmt.Errorf("failed to set key expiration: %w", err)
	}

	if !success {
		return ErrKeyNotFound
	}

	c.logger.Debug("Cache key expiration set successfully",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
	)
	return nil
}

// TTL retorna o tempo de vida restante de uma chave
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to get cache key TTL",
			zap.String("key", key),
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to get key TTL: %w", err)
	}

	return ttl, nil
}

// FlushDB limpa todos os dados do banco atual
func (c *Client) FlushDB(ctx context.Context) error {
	err := c.client.FlushDB(ctx).Err()
	if err != nil {
		c.logger.Error("Failed to flush cache database", zap.Error(err))
		return fmt.Errorf("failed to flush cache database: %w", err)
	}

	c.logger.Info("Cache database flushed successfully")
	return nil
}

// Keys retorna todas as chaves que correspondem a um padrão
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Error("Failed to get cache keys by pattern",
			zap.String("pattern", pattern),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get keys by pattern: %w", err)
	}

	c.logger.Debug("Cache keys retrieved by pattern",
		zap.String("pattern", pattern),
		zap.Int("count", len(keys)),
	)
	return keys, nil
}
