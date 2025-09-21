package cache

import (
	"context"
	"time"
)

// Cache define a interface para operações de cache
type Cache interface {
	// Set armazena um valor no cache com TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get recupera um valor do cache
	Get(ctx context.Context, key string, dest interface{}) error

	// Delete remove um ou mais valores do cache
	Delete(ctx context.Context, keys ...string) error

	// Exists verifica se uma ou mais chaves existem no cache
	Exists(ctx context.Context, keys ...string) (int64, error)

	// SetNX armazena um valor apenas se a chave não existir
	SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)

	// Increment incrementa um valor numérico no cache
	Increment(ctx context.Context, key string) (int64, error)

	// IncrementBy incrementa um valor numérico por uma quantidade específica
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)

	// Expire define um TTL para uma chave existente
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// TTL retorna o tempo de vida restante de uma chave
	TTL(ctx context.Context, key string) (time.Duration, error)

	// FlushDB limpa todos os dados do cache
	FlushDB(ctx context.Context) error

	// Keys retorna todas as chaves que correspondem a um padrão
	Keys(ctx context.Context, pattern string) ([]string, error)

	// Ping testa a conexão com o cache
	Ping(ctx context.Context) error

	// Close fecha a conexão com o cache
	Close() error
}

// CacheManager gerencia múltiplas instâncias de cache
type CacheManager interface {
	// GetCache retorna uma instância de cache por nome
	GetCache(name string) Cache

	// SetCache define uma instância de cache
	SetCache(name string, cache Cache)

	// RemoveCache remove uma instância de cache
	RemoveCache(name string)

	// Close fecha todas as conexões de cache
	Close() error
}

// CacheConfig contém configurações genéricas de cache
type CacheConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`
	DefaultTTL time.Duration `json:"default_ttl" yaml:"default_ttl"`
	KeyPrefix  string        `json:"key_prefix" yaml:"key_prefix"`
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"`
}

// KeyBuilder constrói chaves de cache padronizadas
type KeyBuilder interface {
	// BuildKey constrói uma chave de cache
	BuildKey(parts ...string) string

	// BuildKeyWithTenant constrói uma chave de cache com tenant
	BuildKeyWithTenant(tenantID string, parts ...string) string

	// BuildKeyWithExpiration constrói uma chave de cache com expiração customizada
	BuildKeyWithExpiration(ttl time.Duration, parts ...string) (string, time.Duration)
}
