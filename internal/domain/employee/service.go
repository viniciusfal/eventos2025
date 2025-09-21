package employee

import (
	"context"
	"time"

	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"go.uber.org/zap"
)

// Service define os serviços de domínio para Employee
type Service interface {
	// CreateEmployee cria um novo funcionário com validações de negócio
	CreateEmployee(ctx context.Context, tenantID value_objects.UUID, fullName, identity, identityType, phone, email string, dateOfBirth *time.Time, createdBy value_objects.UUID) (*Employee, error)

	// UpdateEmployee atualiza um funcionário existente
	UpdateEmployee(ctx context.Context, id value_objects.UUID, fullName, identity, identityType, phone, email string, dateOfBirth *time.Time, updatedBy value_objects.UUID) (*Employee, error)

	// UpdateEmployeePhoto atualiza a foto de um funcionário
	UpdateEmployeePhoto(ctx context.Context, id value_objects.UUID, photoURL string, updatedBy value_objects.UUID) error

	// UpdateEmployeeFaceEmbedding atualiza o embedding facial de um funcionário
	UpdateEmployeeFaceEmbedding(ctx context.Context, id value_objects.UUID, embedding []float32, updatedBy value_objects.UUID) error

	// GetEmployee busca um funcionário pelo ID
	GetEmployee(ctx context.Context, id value_objects.UUID) (*Employee, error)

	// GetEmployeeByTenant busca um funcionário pelo ID dentro de um tenant
	GetEmployeeByTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Employee, error)

	// GetEmployeeByIdentity busca um funcionário pela identidade
	GetEmployeeByIdentity(ctx context.Context, identity string, tenantID *value_objects.UUID) (*Employee, error)

	// GetEmployeeByEmail busca um funcionário pelo email
	GetEmployeeByEmail(ctx context.Context, email string, tenantID *value_objects.UUID) (*Employee, error)

	// ActivateEmployee ativa um funcionário
	ActivateEmployee(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// DeactivateEmployee desativa um funcionário
	DeactivateEmployee(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error

	// ListEmployees lista funcionários com filtros
	ListEmployees(ctx context.Context, filters ListFilters) ([]*Employee, int, error)

	// ListEmployeesByTenant lista funcionários de um tenant específico
	ListEmployeesByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Employee, int, error)

	// ListEmployeesByPartner lista funcionários de um parceiro específico
	ListEmployeesByPartner(ctx context.Context, partnerID value_objects.UUID, filters ListFilters) ([]*Employee, int, error)

	// ListEmployeesByEvent lista funcionários associados a um evento
	ListEmployeesByEvent(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Employee, int, error)

	// DeleteEmployee remove um funcionário (soft delete)
	DeleteEmployee(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error

	// RecognizeFace reconhece um funcionário por embedding facial
	RecognizeFace(ctx context.Context, embedding []float32, tenantID *value_objects.UUID, threshold float32) ([]*FaceRecognitionResult, error)

	// ValidateEmployeeForCheckIn valida se um funcionário pode fazer check-in
	ValidateEmployeeForCheckIn(ctx context.Context, employeeID value_objects.UUID) (*Employee, error)

	// ValidateEmployeeForCheckOut valida se um funcionário pode fazer check-out
	ValidateEmployeeForCheckOut(ctx context.Context, employeeID value_objects.UUID) (*Employee, error)

	// GetEmployeesWithFaceEmbedding busca funcionários que têm embedding facial
	GetEmployeesWithFaceEmbedding(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Employee, int, error)
}

// DomainService implementa os serviços de domínio para Employee
type DomainService struct {
	repository Repository
	logger     *zap.Logger
}

// NewDomainService cria uma nova instância do serviço de domínio
func NewDomainService(repository Repository, logger *zap.Logger) Service {
	return &DomainService{
		repository: repository,
		logger:     logger,
	}
}

// CreateEmployee cria um novo funcionário com validações de negócio
func (s *DomainService) CreateEmployee(ctx context.Context, tenantID value_objects.UUID, fullName, identity, identityType, phone, email string, dateOfBirth *time.Time, createdBy value_objects.UUID) (*Employee, error) {
	s.logger.Debug("Creating new employee",
		zap.String("tenant_id", tenantID.String()),
		zap.String("full_name", fullName),
		zap.String("identity", identity),
		zap.String("email", email),
		zap.String("created_by", createdBy.String()),
	)

	// Verificar se já existe funcionário com a mesma identidade no tenant (se identidade fornecida)
	if identity != "" {
		exists, err := s.repository.ExistsByIdentityInTenant(ctx, identity, tenantID, nil)
		if err != nil {
			s.logger.Error("Failed to check identity uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate identity uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("employee", "identity", identity)
		}
	}

	// Verificar se já existe funcionário com o mesmo email no tenant (se email fornecido)
	if email != "" {
		exists, err := s.repository.ExistsByEmailInTenant(ctx, email, tenantID, nil)
		if err != nil {
			s.logger.Error("Failed to check email uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate email uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("employee", "email", email)
		}
	}

	// Criar nova instância do funcionário
	employee, err := NewEmployee(tenantID, fullName, identity, identityType, phone, email, dateOfBirth, createdBy)
	if err != nil {
		s.logger.Error("Failed to create employee instance", zap.Error(err))
		return nil, err
	}

	// Persistir no repositório
	if err := s.repository.Create(ctx, employee); err != nil {
		s.logger.Error("Failed to persist employee", zap.Error(err))
		return nil, errors.NewInternalError("failed to create employee", err)
	}

	s.logger.Info("Employee created successfully",
		zap.String("employee_id", employee.ID.String()),
		zap.String("full_name", employee.FullName),
		zap.String("tenant_id", employee.TenantID.String()),
	)

	return employee, nil
}

// UpdateEmployee atualiza um funcionário existente
func (s *DomainService) UpdateEmployee(ctx context.Context, id value_objects.UUID, fullName, identity, identityType, phone, email string, dateOfBirth *time.Time, updatedBy value_objects.UUID) (*Employee, error) {
	s.logger.Debug("Updating employee",
		zap.String("employee_id", id.String()),
		zap.String("full_name", fullName),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar funcionário existente
	employee, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get employee for update", zap.Error(err))
		return nil, errors.NewInternalError("failed to get employee", err)
	}
	if employee == nil {
		return nil, errors.NewNotFoundError("employee", id.String())
	}

	// Verificar unicidade da identidade no tenant (se alterada)
	if identity != "" && identity != employee.Identity {
		exists, err := s.repository.ExistsByIdentityInTenant(ctx, identity, employee.TenantID, &id)
		if err != nil {
			s.logger.Error("Failed to check identity uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate identity uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("employee", "identity", identity)
		}
	}

	// Verificar unicidade do email no tenant (se alterado)
	if email != "" && email != employee.Email {
		exists, err := s.repository.ExistsByEmailInTenant(ctx, email, employee.TenantID, &id)
		if err != nil {
			s.logger.Error("Failed to check email uniqueness in tenant", zap.Error(err))
			return nil, errors.NewInternalError("failed to validate email uniqueness", err)
		}
		if exists {
			return nil, errors.NewAlreadyExistsError("employee", "email", email)
		}
	}

	// Atualizar dados do funcionário
	if err := employee.Update(fullName, identity, identityType, phone, email, dateOfBirth, updatedBy); err != nil {
		s.logger.Error("Failed to update employee data", zap.Error(err))
		return nil, err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, employee); err != nil {
		s.logger.Error("Failed to persist employee update", zap.Error(err))
		return nil, errors.NewInternalError("failed to update employee", err)
	}

	s.logger.Info("Employee updated successfully",
		zap.String("employee_id", employee.ID.String()),
	)

	return employee, nil
}

// UpdateEmployeePhoto atualiza a foto de um funcionário
func (s *DomainService) UpdateEmployeePhoto(ctx context.Context, id value_objects.UUID, photoURL string, updatedBy value_objects.UUID) error {
	s.logger.Debug("Updating employee photo",
		zap.String("employee_id", id.String()),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar funcionário existente
	employee, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get employee for photo update", zap.Error(err))
		return errors.NewInternalError("failed to get employee", err)
	}
	if employee == nil {
		return errors.NewNotFoundError("employee", id.String())
	}

	// Atualizar foto
	if err := employee.UpdatePhoto(photoURL, updatedBy); err != nil {
		s.logger.Error("Failed to update employee photo", zap.Error(err))
		return err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, employee); err != nil {
		s.logger.Error("Failed to persist employee photo update", zap.Error(err))
		return errors.NewInternalError("failed to update employee photo", err)
	}

	s.logger.Info("Employee photo updated successfully",
		zap.String("employee_id", id.String()),
	)

	return nil
}

// UpdateEmployeeFaceEmbedding atualiza o embedding facial de um funcionário
func (s *DomainService) UpdateEmployeeFaceEmbedding(ctx context.Context, id value_objects.UUID, embedding []float32, updatedBy value_objects.UUID) error {
	s.logger.Debug("Updating employee face embedding",
		zap.String("employee_id", id.String()),
		zap.Int("embedding_size", len(embedding)),
		zap.String("updated_by", updatedBy.String()),
	)

	// Buscar funcionário existente
	employee, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get employee for face embedding update", zap.Error(err))
		return errors.NewInternalError("failed to get employee", err)
	}
	if employee == nil {
		return errors.NewNotFoundError("employee", id.String())
	}

	// Atualizar embedding facial
	if err := employee.UpdateFaceEmbedding(embedding, updatedBy); err != nil {
		s.logger.Error("Failed to update employee face embedding", zap.Error(err))
		return err
	}

	// Persistir alterações
	if err := s.repository.Update(ctx, employee); err != nil {
		s.logger.Error("Failed to persist employee face embedding update", zap.Error(err))
		return errors.NewInternalError("failed to update employee face embedding", err)
	}

	s.logger.Info("Employee face embedding updated successfully",
		zap.String("employee_id", id.String()),
	)

	return nil
}

// GetEmployee busca um funcionário pelo ID
func (s *DomainService) GetEmployee(ctx context.Context, id value_objects.UUID) (*Employee, error) {
	employee, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get employee", zap.Error(err))
		return nil, errors.NewInternalError("failed to get employee", err)
	}
	if employee == nil {
		return nil, errors.NewNotFoundError("employee", id.String())
	}

	return employee, nil
}

// GetEmployeeByTenant busca um funcionário pelo ID dentro de um tenant
func (s *DomainService) GetEmployeeByTenant(ctx context.Context, id, tenantID value_objects.UUID) (*Employee, error) {
	employee, err := s.repository.GetByIDAndTenant(ctx, id, tenantID)
	if err != nil {
		s.logger.Error("Failed to get employee by tenant", zap.Error(err))
		return nil, errors.NewInternalError("failed to get employee", err)
	}
	if employee == nil {
		return nil, errors.NewNotFoundError("employee", id.String())
	}

	return employee, nil
}

// GetEmployeeByIdentity busca um funcionário pela identidade
func (s *DomainService) GetEmployeeByIdentity(ctx context.Context, identity string, tenantID *value_objects.UUID) (*Employee, error) {
	var employee *Employee
	var err error

	if tenantID != nil {
		employee, err = s.repository.GetByIdentityAndTenant(ctx, identity, *tenantID)
	} else {
		employee, err = s.repository.GetByIdentity(ctx, identity)
	}

	if err != nil {
		s.logger.Error("Failed to get employee by identity", zap.Error(err))
		return nil, errors.NewInternalError("failed to get employee", err)
	}
	if employee == nil {
		return nil, errors.NewNotFoundError("employee", identity)
	}

	return employee, nil
}

// GetEmployeeByEmail busca um funcionário pelo email
func (s *DomainService) GetEmployeeByEmail(ctx context.Context, email string, tenantID *value_objects.UUID) (*Employee, error) {
	var employee *Employee
	var err error

	if tenantID != nil {
		employee, err = s.repository.GetByEmailAndTenant(ctx, email, *tenantID)
	} else {
		employee, err = s.repository.GetByEmail(ctx, email)
	}

	if err != nil {
		s.logger.Error("Failed to get employee by email", zap.Error(err))
		return nil, errors.NewInternalError("failed to get employee", err)
	}
	if employee == nil {
		return nil, errors.NewNotFoundError("employee", email)
	}

	return employee, nil
}

// ActivateEmployee ativa um funcionário
func (s *DomainService) ActivateEmployee(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	employee, err := s.GetEmployee(ctx, id)
	if err != nil {
		return err
	}

	if employee.IsActive() {
		return errors.NewValidationError("status", "employee is already active")
	}

	employee.Activate(updatedBy)

	if err := s.repository.Update(ctx, employee); err != nil {
		s.logger.Error("Failed to activate employee", zap.Error(err))
		return errors.NewInternalError("failed to activate employee", err)
	}

	s.logger.Info("Employee activated successfully",
		zap.String("employee_id", id.String()),
	)

	return nil
}

// DeactivateEmployee desativa um funcionário
func (s *DomainService) DeactivateEmployee(ctx context.Context, id value_objects.UUID, updatedBy value_objects.UUID) error {
	employee, err := s.GetEmployee(ctx, id)
	if err != nil {
		return err
	}

	if !employee.IsActive() {
		return errors.NewValidationError("status", "employee is already inactive")
	}

	employee.Deactivate(updatedBy)

	if err := s.repository.Update(ctx, employee); err != nil {
		s.logger.Error("Failed to deactivate employee", zap.Error(err))
		return errors.NewInternalError("failed to deactivate employee", err)
	}

	s.logger.Info("Employee deactivated successfully",
		zap.String("employee_id", id.String()),
	)

	return nil
}

// ListEmployees lista funcionários com filtros
func (s *DomainService) ListEmployees(ctx context.Context, filters ListFilters) ([]*Employee, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	employees, total, err := s.repository.List(ctx, filters)
	if err != nil {
		s.logger.Error("Failed to list employees", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list employees", err)
	}

	return employees, total, nil
}

// ListEmployeesByTenant lista funcionários de um tenant específico
func (s *DomainService) ListEmployeesByTenant(ctx context.Context, tenantID value_objects.UUID, filters ListFilters) ([]*Employee, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	employees, total, err := s.repository.ListByTenant(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to list employees by tenant", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list employees", err)
	}

	return employees, total, nil
}

// ListEmployeesByPartner lista funcionários de um parceiro específico
func (s *DomainService) ListEmployeesByPartner(ctx context.Context, partnerID value_objects.UUID, filters ListFilters) ([]*Employee, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	employees, total, err := s.repository.ListByPartner(ctx, partnerID, filters)
	if err != nil {
		s.logger.Error("Failed to list employees by partner", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list employees by partner", err)
	}

	return employees, total, nil
}

// ListEmployeesByEvent lista funcionários associados a um evento
func (s *DomainService) ListEmployeesByEvent(ctx context.Context, eventID value_objects.UUID, filters ListFilters) ([]*Employee, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	employees, total, err := s.repository.ListByEvent(ctx, eventID, filters)
	if err != nil {
		s.logger.Error("Failed to list employees by event", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list employees by event", err)
	}

	return employees, total, nil
}

// DeleteEmployee remove um funcionário (soft delete)
func (s *DomainService) DeleteEmployee(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	// Verificar se o funcionário existe
	employee, err := s.GetEmployee(ctx, id)
	if err != nil {
		return err
	}

	// Verificar se o funcionário pode ser removido (regras de negócio)
	if employee.IsActive() {
		return errors.NewValidationError("status", "cannot delete active employee")
	}

	if err := s.repository.Delete(ctx, id, deletedBy); err != nil {
		s.logger.Error("Failed to delete employee", zap.Error(err))
		return errors.NewInternalError("failed to delete employee", err)
	}

	s.logger.Info("Employee deleted successfully",
		zap.String("employee_id", id.String()),
	)

	return nil
}

// RecognizeFace reconhece um funcionário por embedding facial
func (s *DomainService) RecognizeFace(ctx context.Context, embedding []float32, tenantID *value_objects.UUID, threshold float32) ([]*FaceRecognitionResult, error) {
	s.logger.Debug("Performing face recognition",
		zap.Int("embedding_size", len(embedding)),
		zap.Float32("threshold", threshold),
	)

	if len(embedding) != 512 {
		return nil, errors.NewValidationError("embedding", "face embedding must have exactly 512 dimensions")
	}

	if threshold < 0.5 || threshold > 1.0 {
		threshold = 0.75 // Threshold padrão
	}

	// Buscar funcionários similares
	employees, similarities, err := s.repository.FindByFaceEmbedding(ctx, embedding, tenantID, threshold, 10)
	if err != nil {
		s.logger.Error("Failed to find employees by face embedding", zap.Error(err))
		return nil, errors.NewInternalError("failed to perform face recognition", err)
	}

	// Criar resultados
	results := make([]*FaceRecognitionResult, len(employees))
	for i, employee := range employees {
		result := &FaceRecognitionResult{
			Employee:   employee,
			Similarity: similarities[i],
		}
		result.Confidence = result.GetConfidenceLevel()
		results[i] = result
	}

	s.logger.Info("Face recognition completed",
		zap.Int("matches_found", len(results)),
	)

	return results, nil
}

// ValidateEmployeeForCheckIn valida se um funcionário pode fazer check-in
func (s *DomainService) ValidateEmployeeForCheckIn(ctx context.Context, employeeID value_objects.UUID) (*Employee, error) {
	employee, err := s.GetEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	if !employee.IsActive() {
		return nil, errors.NewValidationError("employee", "employee is not active")
	}

	return employee, nil
}

// ValidateEmployeeForCheckOut valida se um funcionário pode fazer check-out
func (s *DomainService) ValidateEmployeeForCheckOut(ctx context.Context, employeeID value_objects.UUID) (*Employee, error) {
	employee, err := s.GetEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	if !employee.IsActive() {
		return nil, errors.NewValidationError("employee", "employee is not active")
	}

	return employee, nil
}

// GetEmployeesWithFaceEmbedding busca funcionários que têm embedding facial
func (s *DomainService) GetEmployeesWithFaceEmbedding(ctx context.Context, tenantID *value_objects.UUID, filters ListFilters) ([]*Employee, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	employees, total, err := s.repository.GetEmployeesWithFaceEmbedding(ctx, tenantID, filters)
	if err != nil {
		s.logger.Error("Failed to get employees with face embedding", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to get employees with face embedding", err)
	}

	return employees, total, nil
}
