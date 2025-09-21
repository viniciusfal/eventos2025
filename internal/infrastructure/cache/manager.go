package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultCacheManager implementa CacheManager
type DefaultCacheManager struct {
	caches map[string]Cache
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewDefaultCacheManager cria uma nova instância do gerenciador de cache
func NewDefaultCacheManager(logger *zap.Logger) *DefaultCacheManager {
	return &DefaultCacheManager{
		caches: make(map[string]Cache),
		logger: logger,
	}
}

// GetCache retorna uma instância de cache por nome
func (cm *DefaultCacheManager) GetCache(name string) Cache {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.caches[name]
}

// SetCache define uma instância de cache
func (cm *DefaultCacheManager) SetCache(name string, cache Cache) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Fechar cache anterior se existir
	if existing, exists := cm.caches[name]; exists {
		if err := existing.Close(); err != nil {
			cm.logger.Warn("Failed to close existing cache",
				zap.String("name", name),
				zap.Error(err),
			)
		}
	}

	cm.caches[name] = cache
	cm.logger.Info("Cache instance set", zap.String("name", name))
}

// RemoveCache remove uma instância de cache
func (cm *DefaultCacheManager) RemoveCache(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cache, exists := cm.caches[name]; exists {
		if err := cache.Close(); err != nil {
			cm.logger.Warn("Failed to close cache during removal",
				zap.String("name", name),
				zap.Error(err),
			)
		}
		delete(cm.caches, name)
		cm.logger.Info("Cache instance removed", zap.String("name", name))
	}
}

// Close fecha todas as conexões de cache
func (cm *DefaultCacheManager) Close() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	var lastErr error
	for name, cache := range cm.caches {
		if err := cache.Close(); err != nil {
			cm.logger.Error("Failed to close cache",
				zap.String("name", name),
				zap.Error(err),
			)
			lastErr = err
		}
	}

	// Limpar o mapa
	cm.caches = make(map[string]Cache)

	if lastErr != nil {
		return fmt.Errorf("errors occurred while closing caches: %w", lastErr)
	}

	cm.logger.Info("All cache instances closed")
	return nil
}

// HealthCheck verifica a saúde de todos os caches
func (cm *DefaultCacheManager) HealthCheck(ctx context.Context) map[string]error {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	results := make(map[string]error)

	for name, cache := range cm.caches {
		if err := cache.Ping(ctx); err != nil {
			results[name] = err
			cm.logger.Error("Cache health check failed",
				zap.String("name", name),
				zap.Error(err),
			)
		} else {
			results[name] = nil
		}
	}

	return results
}

// GetCacheNames retorna os nomes de todos os caches registrados
func (cm *DefaultCacheManager) GetCacheNames() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	names := make([]string, 0, len(cm.caches))
	for name := range cm.caches {
		names = append(names, name)
	}

	return names
}

// CacheService fornece operações de cache de alto nível
type CacheService struct {
	manager    CacheManager
	keyBuilder KeyBuilder
	logger     *zap.Logger
	defaultTTL time.Duration
}

// NewCacheService cria uma nova instância do serviço de cache
func NewCacheService(manager CacheManager, keyBuilder KeyBuilder, logger *zap.Logger, defaultTTL time.Duration) *CacheService {
	return &CacheService{
		manager:    manager,
		keyBuilder: keyBuilder,
		logger:     logger,
		defaultTTL: defaultTTL,
	}
}

// GetDefaultCache retorna o cache padrão
func (cs *CacheService) GetDefaultCache() Cache {
	return cs.manager.GetCache("default")
}

// CacheEntity armazena uma entidade no cache
func (cs *CacheService) CacheEntity(ctx context.Context, cacheName, tenantID, entityType, entityID string, entity interface{}, ttl time.Duration) error {
	cache := cs.manager.GetCache(cacheName)
	if cache == nil {
		return fmt.Errorf("cache '%s' not found", cacheName)
	}

	key := cs.keyBuilder.BuildKeyWithTenant(tenantID, entityType, entityID)
	if ttl <= 0 {
		ttl = cs.defaultTTL
	}

	err := cache.Set(ctx, key, entity, ttl)
	if err != nil {
		cs.logger.Error("Failed to cache entity",
			zap.String("cache", cacheName),
			zap.String("key", key),
			zap.String("entity_type", entityType),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("Entity cached successfully",
		zap.String("cache", cacheName),
		zap.String("key", key),
		zap.String("entity_type", entityType),
		zap.Duration("ttl", ttl),
	)
	return nil
}

// GetEntity recupera uma entidade do cache
func (cs *CacheService) GetEntity(ctx context.Context, cacheName, tenantID, entityType, entityID string, dest interface{}) error {
	cache := cs.manager.GetCache(cacheName)
	if cache == nil {
		return fmt.Errorf("cache '%s' not found", cacheName)
	}

	key := cs.keyBuilder.BuildKeyWithTenant(tenantID, entityType, entityID)
	err := cache.Get(ctx, key, dest)
	if err != nil {
		cs.logger.Debug("Cache miss for entity",
			zap.String("cache", cacheName),
			zap.String("key", key),
			zap.String("entity_type", entityType),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("Entity retrieved from cache",
		zap.String("cache", cacheName),
		zap.String("key", key),
		zap.String("entity_type", entityType),
	)
	return nil
}

// InvalidateEntity remove uma entidade do cache
func (cs *CacheService) InvalidateEntity(ctx context.Context, cacheName, tenantID, entityType, entityID string) error {
	cache := cs.manager.GetCache(cacheName)
	if cache == nil {
		return fmt.Errorf("cache '%s' not found", cacheName)
	}

	key := cs.keyBuilder.BuildKeyWithTenant(tenantID, entityType, entityID)
	err := cache.Delete(ctx, key)
	if err != nil {
		cs.logger.Error("Failed to invalidate entity cache",
			zap.String("cache", cacheName),
			zap.String("key", key),
			zap.String("entity_type", entityType),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Debug("Entity cache invalidated",
		zap.String("cache", cacheName),
		zap.String("key", key),
		zap.String("entity_type", entityType),
	)
	return nil
}

// InvalidatePattern remove todas as chaves que correspondem a um padrão
func (cs *CacheService) InvalidatePattern(ctx context.Context, cacheName, pattern string) error {
	cache := cs.manager.GetCache(cacheName)
	if cache == nil {
		return fmt.Errorf("cache '%s' not found", cacheName)
	}

	keys, err := cache.Keys(ctx, pattern)
	if err != nil {
		cs.logger.Error("Failed to get keys for pattern invalidation",
			zap.String("cache", cacheName),
			zap.String("pattern", pattern),
			zap.Error(err),
		)
		return err
	}

	if len(keys) == 0 {
		cs.logger.Debug("No keys found for pattern",
			zap.String("cache", cacheName),
			zap.String("pattern", pattern),
		)
		return nil
	}

	err = cache.Delete(ctx, keys...)
	if err != nil {
		cs.logger.Error("Failed to invalidate keys by pattern",
			zap.String("cache", cacheName),
			zap.String("pattern", pattern),
			zap.Int("key_count", len(keys)),
			zap.Error(err),
		)
		return err
	}

	cs.logger.Info("Keys invalidated by pattern",
		zap.String("cache", cacheName),
		zap.String("pattern", pattern),
		zap.Int("key_count", len(keys)),
	)
	return nil
}

// Close fecha o serviço de cache
func (cs *CacheService) Close() error {
	return cs.manager.Close()
}
