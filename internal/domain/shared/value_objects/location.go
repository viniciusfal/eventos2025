package value_objects

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Location representa uma coordenada geográfica (latitude, longitude)
type Location struct {
	Latitude  float64
	Longitude float64
}

// NewLocation cria uma nova localização
func NewLocation(latitude, longitude float64) (Location, error) {
	if err := validateCoordinates(latitude, longitude); err != nil {
		return Location{}, err
	}

	return Location{
		Latitude:  latitude,
		Longitude: longitude,
	}, nil
}

// String retorna a representação em string da localização
func (l Location) String() string {
	return fmt.Sprintf("POINT(%f %f)", l.Longitude, l.Latitude)
}

// IsZero verifica se a localização é zero
func (l Location) IsZero() bool {
	return l.Latitude == 0 && l.Longitude == 0
}

// DistanceTo calcula a distância em metros para outra localização usando a fórmula de Haversine
func (l Location) DistanceTo(other Location) float64 {
	const earthRadius = 6371000 // metros

	lat1Rad := l.Latitude * (math.Pi / 180)
	lat2Rad := other.Latitude * (math.Pi / 180)
	deltaLatRad := (other.Latitude - l.Latitude) * (math.Pi / 180)
	deltaLonRad := (other.Longitude - l.Longitude) * (math.Pi / 180)

	a := math.Sin(deltaLatRad/2)*math.Sin(deltaLatRad/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLonRad/2)*math.Sin(deltaLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Value implementa driver.Valuer para persistência no banco (PostGIS)
func (l Location) Value() (driver.Value, error) {
	if l.IsZero() {
		return nil, nil
	}
	return l.String(), nil
}

// Scan implementa sql.Scanner para leitura do banco (PostGIS)
func (l *Location) Scan(value interface{}) error {
	if value == nil {
		l.Latitude = 0
		l.Longitude = 0
		return nil
	}

	switch v := value.(type) {
	case string:
		return l.parseFromString(v)
	case []byte:
		return l.parseFromString(string(v))
	default:
		return fmt.Errorf("cannot scan %T into Location", value)
	}
}

// parseFromString analisa uma string no formato "POINT(lon lat)" ou similar
func (l *Location) parseFromString(s string) error {
	// Remover prefixos comuns do PostGIS
	s = strings.TrimPrefix(s, "POINT(")
	s = strings.TrimSuffix(s, ")")
	s = strings.TrimSpace(s)

	// Dividir por espaço
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return fmt.Errorf("invalid location format: %s", s)
	}

	// Parse longitude (primeiro valor)
	lon, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid longitude: %s", parts[0])
	}

	// Parse latitude (segundo valor)
	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid latitude: %s", parts[1])
	}

	if err := validateCoordinates(lat, lon); err != nil {
		return err
	}

	l.Latitude = lat
	l.Longitude = lon
	return nil
}

// validateCoordinates valida se as coordenadas são válidas
func validateCoordinates(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return fmt.Errorf("invalid latitude: %f (must be between -90 and 90)", latitude)
	}

	if longitude < -180 || longitude > 180 {
		return fmt.Errorf("invalid longitude: %f (must be between -180 and 180)", longitude)
	}

	return nil
}
