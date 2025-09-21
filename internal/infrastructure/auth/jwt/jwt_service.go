package jwt

import (
	"fmt"
	"time"

	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/golang-jwt/jwt/v5"
)

// Claims representa as claims do JWT
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// Service define as operações do serviço JWT
type Service interface {
	// GenerateToken gera um token JWT para o usuário
	GenerateToken(userID, tenantID value_objects.UUID, username, email string) (string, error)

	// GenerateRefreshToken gera um refresh token
	GenerateRefreshToken(userID, tenantID value_objects.UUID, username, email string) (string, error)

	// ValidateToken valida um token JWT e retorna as claims
	ValidateToken(tokenString string) (*Claims, error)

	// ValidateRefreshToken valida um refresh token
	ValidateRefreshToken(tokenString string) (*Claims, error)

	// RefreshToken gera um novo token a partir de um refresh token válido
	RefreshToken(refreshTokenString string) (string, string, error)
}

// JWTService implementa o serviço JWT
type JWTService struct {
	secretKey         []byte
	expiration        time.Duration
	refreshExpiration time.Duration
	issuer            string
}

// Config representa a configuração do JWT
type Config struct {
	SecretKey         string
	Expiration        time.Duration
	RefreshExpiration time.Duration
	Issuer            string
}

// NewJWTService cria uma nova instância do serviço JWT
func NewJWTService(config Config) Service {
	return &JWTService{
		secretKey:         []byte(config.SecretKey),
		expiration:        config.Expiration,
		refreshExpiration: config.RefreshExpiration,
		issuer:            config.Issuer,
	}
}

// GenerateToken gera um token JWT para o usuário
func (s *JWTService) GenerateToken(userID, tenantID value_objects.UUID, username, email string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.expiration)

	claims := &Claims{
		UserID:   userID.String(),
		TenantID: tenantID.String(),
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken gera um refresh token
func (s *JWTService) GenerateRefreshToken(userID, tenantID value_objects.UUID, username, email string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.refreshExpiration)

	claims := &Claims{
		UserID:   userID.String(),
		TenantID: tenantID.String(),
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken valida um token JWT e retorna as claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Verificar se o token não expirou
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// ValidateRefreshToken valida um refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	// A validação é a mesma, mas podemos adicionar lógicas específicas para refresh tokens
	return s.ValidateToken(tokenString)
}

// RefreshToken gera um novo token a partir de um refresh token válido
func (s *JWTService) RefreshToken(refreshTokenString string) (string, string, error) {
	// Validar o refresh token
	claims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Parse dos UUIDs
	userID, err := value_objects.ParseUUID(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("invalid user ID in token: %w", err)
	}

	tenantID, err := value_objects.ParseUUID(claims.TenantID)
	if err != nil {
		return "", "", fmt.Errorf("invalid tenant ID in token: %w", err)
	}

	// Gerar novo access token
	newToken, err := s.GenerateToken(userID, tenantID, claims.Username, claims.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new token: %w", err)
	}

	// Gerar novo refresh token
	newRefreshToken, err := s.GenerateRefreshToken(userID, tenantID, claims.Username, claims.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	return newToken, newRefreshToken, nil
}

// GetUserIDFromToken extrai o ID do usuário do token
func (s *JWTService) GetUserIDFromToken(tokenString string) (value_objects.UUID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return value_objects.UUID{}, err
	}

	return value_objects.ParseUUID(claims.UserID)
}

// GetTenantIDFromToken extrai o ID do tenant do token
func (s *JWTService) GetTenantIDFromToken(tokenString string) (value_objects.UUID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return value_objects.UUID{}, err
	}

	return value_objects.ParseUUID(claims.TenantID)
}
