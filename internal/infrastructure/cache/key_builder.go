package cache

import (
	"fmt"
	"strings"
	"time"
)

// DefaultKeyBuilder implementa KeyBuilder com funcionalidades padrão
type DefaultKeyBuilder struct {
	prefix     string
	separator  string
	defaultTTL time.Duration
}

// NewDefaultKeyBuilder cria uma nova instância do construtor de chaves
func NewDefaultKeyBuilder(prefix string, defaultTTL time.Duration) *DefaultKeyBuilder {
	return &DefaultKeyBuilder{
		prefix:     prefix,
		separator:  ":",
		defaultTTL: defaultTTL,
	}
}

// BuildKey constrói uma chave de cache
func (kb *DefaultKeyBuilder) BuildKey(parts ...string) string {
	if len(parts) == 0 {
		return kb.prefix
	}

	// Filtrar partes vazias
	validParts := make([]string, 0, len(parts)+1)
	if kb.prefix != "" {
		validParts = append(validParts, kb.prefix)
	}

	for _, part := range parts {
		if part != "" {
			validParts = append(validParts, part)
		}
	}

	return strings.Join(validParts, kb.separator)
}

// BuildKeyWithTenant constrói uma chave de cache com tenant
func (kb *DefaultKeyBuilder) BuildKeyWithTenant(tenantID string, parts ...string) string {
	if tenantID == "" {
		return kb.BuildKey(parts...)
	}

	// Adicionar tenant como primeiro elemento após o prefixo
	tenantParts := make([]string, 0, len(parts)+2)
	tenantParts = append(tenantParts, "tenant", tenantID)
	tenantParts = append(tenantParts, parts...)

	return kb.BuildKey(tenantParts...)
}

// BuildKeyWithExpiration constrói uma chave de cache com expiração customizada
func (kb *DefaultKeyBuilder) BuildKeyWithExpiration(ttl time.Duration, parts ...string) (string, time.Duration) {
	key := kb.BuildKey(parts...)
	if ttl <= 0 {
		ttl = kb.defaultTTL
	}
	return key, ttl
}

// Métodos de conveniência para tipos específicos de chaves

// UserKey constrói chaves relacionadas a usuários
func (kb *DefaultKeyBuilder) UserKey(tenantID, userID string, suffix ...string) string {
	parts := []string{"user", userID}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// RoleKey constrói chaves relacionadas a roles
func (kb *DefaultKeyBuilder) RoleKey(tenantID, roleID string, suffix ...string) string {
	parts := []string{"role", roleID}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// PermissionKey constrói chaves relacionadas a permissões
func (kb *DefaultKeyBuilder) PermissionKey(tenantID, permissionID string, suffix ...string) string {
	parts := []string{"permission", permissionID}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// EventKey constrói chaves relacionadas a eventos
func (kb *DefaultKeyBuilder) EventKey(tenantID, eventID string, suffix ...string) string {
	parts := []string{"event", eventID}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// EmployeeKey constrói chaves relacionadas a funcionários
func (kb *DefaultKeyBuilder) EmployeeKey(tenantID, employeeID string, suffix ...string) string {
	parts := []string{"employee", employeeID}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// PartnerKey constrói chaves relacionadas a parceiros
func (kb *DefaultKeyBuilder) PartnerKey(tenantID, partnerID string, suffix ...string) string {
	parts := []string{"partner", partnerID}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// CheckinKey constrói chaves relacionadas a check-ins
func (kb *DefaultKeyBuilder) CheckinKey(tenantID, checkinID string, suffix ...string) string {
	parts := []string{"checkin", checkinID}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// CheckoutKey constrói chaves relacionadas a check-outs
func (kb *DefaultKeyBuilder) CheckoutKey(tenantID, checkoutID string, suffix ...string) string {
	parts := []string{"checkout", checkoutID}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// SessionKey constrói chaves relacionadas a sessões
func (kb *DefaultKeyBuilder) SessionKey(sessionID string, suffix ...string) string {
	parts := []string{"session", sessionID}
	parts = append(parts, suffix...)
	return kb.BuildKey(parts...)
}

// TokenKey constrói chaves relacionadas a tokens
func (kb *DefaultKeyBuilder) TokenKey(tokenType, tokenID string, suffix ...string) string {
	parts := []string{"token", tokenType, tokenID}
	parts = append(parts, suffix...)
	return kb.BuildKey(parts...)
}

// StatsKey constrói chaves relacionadas a estatísticas
func (kb *DefaultKeyBuilder) StatsKey(tenantID, statsType string, suffix ...string) string {
	parts := []string{"stats", statsType}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// ListKey constrói chaves relacionadas a listas paginadas
func (kb *DefaultKeyBuilder) ListKey(tenantID, entityType string, page, pageSize int, filters ...string) string {
	parts := []string{"list", entityType, fmt.Sprintf("page:%d", page), fmt.Sprintf("size:%d", pageSize)}

	// Adicionar filtros se fornecidos
	if len(filters) > 0 {
		filterStr := strings.Join(filters, ",")
		parts = append(parts, fmt.Sprintf("filters:%s", filterStr))
	}

	return kb.BuildKeyWithTenant(tenantID, parts...)
}

// SearchKey constrói chaves relacionadas a buscas
func (kb *DefaultKeyBuilder) SearchKey(tenantID, entityType, query string, suffix ...string) string {
	parts := []string{"search", entityType, fmt.Sprintf("q:%s", query)}
	parts = append(parts, suffix...)
	return kb.BuildKeyWithTenant(tenantID, parts...)
}
