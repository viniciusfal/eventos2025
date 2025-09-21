package event

import (
	"testing"
	"time"

	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// EventTestSuite é a suíte de testes para Event
type EventTestSuite struct {
	suite.Suite
}

func TestEventSuite(t *testing.T) {
	suite.Run(t, new(EventTestSuite))
}

func (suite *EventTestSuite) TestNewEvent_ValidData() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	name := "Conferência Tech 2024"
	location := "São Paulo, SP"
	fenceEvent := []value_objects.Location{
		{Latitude: -23.5505, Longitude: -46.6333},
		{Latitude: -23.5506, Longitude: -46.6333},
		{Latitude: -23.5506, Longitude: -46.6334},
		{Latitude: -23.5505, Longitude: -46.6334},
	}
	initialDate := time.Now().UTC().Add(24 * time.Hour) // Amanhã
	finalDate := initialDate.Add(8 * time.Hour)         // 8 horas depois

	// Act
	event, err := NewEvent(tenantID, name, location, fenceEvent, initialDate, finalDate, createdBy)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), event)
	assert.NotEmpty(suite.T(), event.ID.String())
	assert.Equal(suite.T(), tenantID, event.TenantID)
	assert.Equal(suite.T(), name, event.Name)
	assert.Equal(suite.T(), location, event.Location)
	assert.NotEmpty(suite.T(), event.FenceEvent) // Deve ter pontos na cerca
	assert.True(suite.T(), event.Active)
	assert.False(suite.T(), event.CreatedAt.IsZero())
	assert.False(suite.T(), event.UpdatedAt.IsZero())
	assert.Equal(suite.T(), event.CreatedAt, event.UpdatedAt)
	assert.NotNil(suite.T(), event.CreatedBy)
	assert.NotNil(suite.T(), event.UpdatedBy)
}

func (suite *EventTestSuite) TestEventActivate() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	initialDate := time.Now().UTC().Add(24 * time.Hour)
	finalDate := initialDate.Add(8 * time.Hour)
	event, _ := NewEvent(tenantID, "Test Event", "Location", []value_objects.Location{}, initialDate, finalDate, createdBy)

	// Verificar estado inicial
	assert.True(suite.T(), event.Active, "Event should be active by default")

	// Desativar primeiro
	event.Deactivate(createdBy)
	assert.False(suite.T(), event.Active, "Event should be inactive after Deactivate")

	// Act
	updatedBy := value_objects.NewUUID()
	event.Activate(updatedBy)

	// Assert
	assert.True(suite.T(), event.Active, "Event should be active after Activate")
	assert.NotNil(suite.T(), event.UpdatedBy)
	assert.False(suite.T(), event.UpdatedAt.IsZero())
	assert.True(suite.T(), !event.UpdatedAt.Before(event.CreatedAt), "UpdatedAt should be after or equal to CreatedAt")
}

func (suite *EventTestSuite) TestEventDeactivate() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	initialDate := time.Now().UTC().Add(24 * time.Hour)
	finalDate := initialDate.Add(8 * time.Hour)
	event, _ := NewEvent(tenantID, "Test Event", "Location", []value_objects.Location{}, initialDate, finalDate, createdBy)

	// Verificar estado inicial
	assert.True(suite.T(), event.Active, "Event should be active by default")

	// Act
	updatedBy := value_objects.NewUUID()
	event.Deactivate(updatedBy)

	// Assert
	assert.False(suite.T(), event.Active, "Event should be inactive after Deactivate")
	assert.NotNil(suite.T(), event.UpdatedBy)
	assert.False(suite.T(), event.UpdatedAt.IsZero())
	assert.True(suite.T(), !event.UpdatedAt.Before(event.CreatedAt), "UpdatedAt should be after or equal to CreatedAt")
}

func (suite *EventTestSuite) TestEventIsActive() {
	// Arrange
	tenantID := value_objects.NewUUID()
	createdBy := value_objects.NewUUID()
	initialDate := time.Now().UTC().Add(24 * time.Hour)
	finalDate := initialDate.Add(8 * time.Hour)
	event, _ := NewEvent(tenantID, "Test Event", "Location", []value_objects.Location{}, initialDate, finalDate, createdBy)

	// Assert
	assert.True(suite.T(), event.IsActive())
}
