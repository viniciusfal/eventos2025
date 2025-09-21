package redis

import "errors"

var (
	// ErrCacheMiss indica que a chave não foi encontrada no cache
	ErrCacheMiss = errors.New("cache miss")

	// ErrKeyNotFound indica que a chave não existe
	ErrKeyNotFound = errors.New("key not found")

	// ErrConnectionFailed indica falha na conexão com o Redis
	ErrConnectionFailed = errors.New("redis connection failed")

	// ErrMarshalFailed indica falha na serialização dos dados
	ErrMarshalFailed = errors.New("failed to marshal data")

	// ErrUnmarshalFailed indica falha na deserialização dos dados
	ErrUnmarshalFailed = errors.New("failed to unmarshal data")
)
