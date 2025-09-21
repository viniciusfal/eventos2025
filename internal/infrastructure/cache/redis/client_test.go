package redis

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// RedisClientTestSuite é a suíte de testes para RedisClient
type RedisClientTestSuite struct {
	suite.Suite
	client *Client
	mini   *miniredis.Miniredis
	logger *zap.Logger
}

// SetupTest é executado antes de cada teste
func (suite *RedisClientTestSuite) SetupTest() {
	// Criar instância do miniredis (Redis em memória para testes)
	mini, err := miniredis.Run()
	suite.Require().NoError(err)

	// Criar logger mock
	suite.logger = zap.NewNop()

	// Criar configuração do Redis
	port, _ := strconv.Atoi(mini.Port())
	config := Config{
		Host:            mini.Host(),
		Port:            port,
		Password:        "",
		DB:              0,
		MaxRetries:      3,
		PoolSize:        10,
		MinIdleConns:    2,
		DialTimeout:     time.Second * 5,
		ReadTimeout:     time.Second * 3,
		WriteTimeout:    time.Second * 3,
		IdleTimeout:     time.Minute * 5,
		ConnMaxLifetime: time.Minute * 10,
	}

	// Criar cliente Redis
	suite.client, err = NewClient(config, suite.logger)
	suite.Require().NoError(err)
	suite.Require().NotNil(suite.client)

	// Guardar referência para cleanup
	suite.mini = mini
}

// TearDownTest é executado após cada teste
func (suite *RedisClientTestSuite) TearDownTest() {
	if suite.client != nil {
		suite.client.Close()
	}
	if suite.mini != nil {
		suite.mini.Close()
	}
}

func TestRedisClientSuite(t *testing.T) {
	suite.Run(t, new(RedisClientTestSuite))
}

func (suite *RedisClientTestSuite) TestNewClient_Success() {
	// Assert
	assert.NotNil(suite.T(), suite.client)
	assert.NotNil(suite.T(), suite.client.client)
}

func (suite *RedisClientTestSuite) TestPing_Success() {
	// Act
	err := suite.client.Ping(context.Background())

	// Assert
	assert.NoError(suite.T(), err)
}

func (suite *RedisClientTestSuite) TestSetAndGet_Success() {
	// Arrange
	ctx := context.Background()
	key := "test_key"
	value := "test_value"
	ttl := time.Minute * 5

	// Act
	err := suite.client.Set(ctx, key, value, ttl)

	// Assert
	assert.NoError(suite.T(), err)

	// Act
	var result string
	err = suite.client.Get(ctx, key, &result)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), value, result)
}

func (suite *RedisClientTestSuite) TestSetAndGet_Struct() {
	// Arrange
	ctx := context.Background()
	key := "test_struct"
	testData := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}{
		Name:  "João Silva",
		Email: "joao@example.com",
		Age:   30,
	}
	ttl := time.Minute * 10

	// Act
	err := suite.client.Set(ctx, key, testData, ttl)

	// Assert
	assert.NoError(suite.T(), err)

	// Act
	var result struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}
	err = suite.client.Get(ctx, key, &result)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testData.Name, result.Name)
	assert.Equal(suite.T(), testData.Email, result.Email)
	assert.Equal(suite.T(), testData.Age, result.Age)
}

func (suite *RedisClientTestSuite) TestGet_NonExistentKey() {
	// Arrange
	ctx := context.Background()
	key := "non_existent_key"

	// Act
	var result string
	err := suite.client.Get(ctx, key, &result)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", result)
	assert.Contains(suite.T(), err.Error(), "cache miss")
}

func (suite *RedisClientTestSuite) TestDelete_Success() {
	// Arrange
	ctx := context.Background()
	key := "test_delete"
	value := "test_value"
	ttl := time.Minute * 5

	// Primeiro, definir um valor
	err := suite.client.Set(ctx, key, value, ttl)
	suite.Require().NoError(err)

	// Act
	err = suite.client.Delete(ctx, key)

	// Assert
	assert.NoError(suite.T(), err)

	// Verificar que a chave foi removida
	var result string
	err = suite.client.Get(ctx, key, &result)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "", result)
}

func (suite *RedisClientTestSuite) TestExists_Success() {
	// Arrange
	ctx := context.Background()
	existingKey := "existing_key"
	nonExistentKey := "non_existent_key"
	value := "test_value"
	ttl := time.Minute * 5

	// Definir um valor para a chave existente
	err := suite.client.Set(ctx, existingKey, value, ttl)
	suite.Require().NoError(err)

	// Act & Assert
	count, err := suite.client.Exists(ctx, existingKey)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)

	count, err = suite.client.Exists(ctx, nonExistentKey)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *RedisClientTestSuite) TestTTL_Success() {
	// Arrange
	ctx := context.Background()
	key := "test_ttl"
	value := "test_value"
	ttl := time.Minute * 5

	// Definir um valor com TTL
	err := suite.client.Set(ctx, key, value, ttl)
	suite.Require().NoError(err)

	// Act
	resultTTL, err := suite.client.TTL(ctx, key)

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), resultTTL > 0)
	assert.True(suite.T(), resultTTL <= ttl)
}

func (suite *RedisClientTestSuite) TestExpire_Success() {
	// Arrange
	ctx := context.Background()
	key := "test_expire"
	value := "test_value"

	// Definir um valor sem TTL (persistente)
	err := suite.client.Set(ctx, key, value, 0)
	suite.Require().NoError(err)

	// Act
	err = suite.client.Expire(ctx, key, time.Minute*10)

	// Assert
	assert.NoError(suite.T(), err)

	// Verificar que agora tem TTL
	resultTTL, err := suite.client.TTL(ctx, key)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), resultTTL > 0)
}

func (suite *RedisClientTestSuite) TestIncrement_Success() {
	// Arrange
	ctx := context.Background()
	key := "test_counter"

	// Act
	result, err := suite.client.Increment(ctx, key)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result)

	// Incrementar novamente
	result, err = suite.client.Increment(ctx, key)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), result)
}

func (suite *RedisClientTestSuite) TestHashOperations_Success() {
	// Arrange
	ctx := context.Background()
	key := "test_hash"
	value := map[string]string{
		"field1": "value1",
		"field2": "value2",
	}
	ttl := time.Minute * 5

	// Act
	err := suite.client.Set(ctx, key, value, ttl)

	// Assert
	assert.NoError(suite.T(), err)

	// Act
	var result map[string]string
	err = suite.client.Get(ctx, key, &result)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), value["field1"], result["field1"])
	assert.Equal(suite.T(), value["field2"], result["field2"])
	assert.Len(suite.T(), result, 2)
}

func (suite *RedisClientTestSuite) TestDeleteMultiple_Success() {
	// Arrange
	ctx := context.Background()

	// Definir alguns dados
	keys := []string{"delete_test_1", "delete_test_2"}
	for _, key := range keys {
		err := suite.client.Set(ctx, key, "value", time.Minute)
		suite.Require().NoError(err)
	}

	// Act
	err := suite.client.Delete(ctx, keys...)

	// Assert
	assert.NoError(suite.T(), err)

	// Verificar que todos os dados foram removidos
	for _, key := range keys {
		var result string
		err := suite.client.Get(ctx, key, &result)
		assert.Error(suite.T(), err)
	}
}
