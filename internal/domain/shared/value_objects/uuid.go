package value_objects

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// UUID representa um identificador único universal
type UUID struct {
	value uuid.UUID
}

// NewUUID cria um novo UUID
func NewUUID() UUID {
	return UUID{value: uuid.New()}
}

// ParseUUID cria um UUID a partir de uma string
func ParseUUID(s string) (UUID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return UUID{}, fmt.Errorf("invalid UUID format: %w", err)
	}
	return UUID{value: parsed}, nil
}

// MustParseUUID cria um UUID a partir de uma string, fazendo panic em caso de erro
func MustParseUUID(s string) UUID {
	u, err := ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// String retorna a representação em string do UUID
func (u UUID) String() string {
	return u.value.String()
}

// IsZero verifica se o UUID é zero
func (u UUID) IsZero() bool {
	return u.value == uuid.Nil
}

// Equals compara dois UUIDs
func (u UUID) Equals(other UUID) bool {
	return u.value == other.value
}

// Value implementa driver.Valuer para persistência no banco
func (u UUID) Value() (driver.Value, error) {
	if u.IsZero() {
		return nil, nil
	}
	return u.value.String(), nil
}

// Scan implementa sql.Scanner para leitura do banco
func (u *UUID) Scan(value interface{}) error {
	if value == nil {
		u.value = uuid.Nil
		return nil
	}

	switch v := value.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("cannot scan %v into UUID: %w", value, err)
		}
		u.value = parsed
	case []byte:
		parsed, err := uuid.ParseBytes(v)
		if err != nil {
			return fmt.Errorf("cannot scan %v into UUID: %w", value, err)
		}
		u.value = parsed
	default:
		return fmt.Errorf("cannot scan %T into UUID", value)
	}

	return nil
}
