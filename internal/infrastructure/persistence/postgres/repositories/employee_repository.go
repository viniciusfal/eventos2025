package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/employee"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// EmployeeRepository implementa a interface employee.Repository usando PostgreSQL
type EmployeeRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewEmployeeRepository cria uma nova instância do repositório de funcionários
func NewEmployeeRepository(db *sqlx.DB, logger *zap.Logger) employee.Repository {
	return &EmployeeRepository{
		db:     db,
		logger: logger,
	}
}

// employeeRow representa uma linha de funcionário no banco de dados
type employeeRow struct {
	ID            string          `db:"id"`
	TenantID      string          `db:"tenant_id"`
	FullName      string          `db:"full_name"`
	Identity      string          `db:"identity"`
	IdentityType  string          `db:"identity_type"`
	DateOfBirth   sql.NullTime    `db:"date_of_birth"`
	PhotoURL      sql.NullString  `db:"photo_url"`
	FaceEmbedding pq.Float32Array `db:"face_embedding"`
	Phone         string          `db:"phone"`
	Email         string          `db:"email"`
	Active        bool            `db:"active"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
	CreatedBy     sql.NullString  `db:"created_by"`
	UpdatedBy     sql.NullString  `db:"updated_by"`
}

// toEntity converte employeeRow para entidade Employee
func (r *employeeRow) toEntity() (*employee.Employee, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, errors.NewDomainError("INVALID_ID", "invalid employee ID", err)
	}

	tenantID, err := value_objects.ParseUUID(r.TenantID)
	if err != nil {
		return nil, errors.NewDomainError("INVALID_TENANT_ID", "invalid tenant ID", err)
	}

	emp := &employee.Employee{
		ID:           id,
		TenantID:     tenantID,
		FullName:     r.FullName,
		Identity:     r.Identity,
		IdentityType: r.IdentityType,
		Phone:        r.Phone,
		Email:        r.Email,
		Active:       r.Active,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}

	if r.DateOfBirth.Valid {
		emp.DateOfBirth = &r.DateOfBirth.Time
	}

	if r.PhotoURL.Valid {
		emp.PhotoURL = r.PhotoURL.String
	}

	if len(r.FaceEmbedding) > 0 {
		emp.FaceEmbedding = []float32(r.FaceEmbedding)
	}

	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err == nil {
			emp.CreatedBy = &createdBy
		}
	}

	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err == nil {
			emp.UpdatedBy = &updatedBy
		}
	}

	return emp, nil
}

// fromEntity converte entidade Employee para employeeRow
func (repo *EmployeeRepository) fromEntity(emp *employee.Employee) *employeeRow {
	row := &employeeRow{
		ID:           emp.ID.String(),
		TenantID:     emp.TenantID.String(),
		FullName:     emp.FullName,
		Identity:     emp.Identity,
		IdentityType: emp.IdentityType,
		Phone:        emp.Phone,
		Email:        emp.Email,
		Active:       emp.Active,
		CreatedAt:    emp.CreatedAt,
		UpdatedAt:    emp.UpdatedAt,
	}

	if emp.DateOfBirth != nil {
		row.DateOfBirth = sql.NullTime{Time: *emp.DateOfBirth, Valid: true}
	}

	if emp.PhotoURL != "" {
		row.PhotoURL = sql.NullString{String: emp.PhotoURL, Valid: true}
	}

	if len(emp.FaceEmbedding) > 0 {
		row.FaceEmbedding = pq.Float32Array(emp.FaceEmbedding)
	}

	if emp.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: emp.CreatedBy.String(), Valid: true}
	}

	if emp.UpdatedBy != nil {
		row.UpdatedBy = sql.NullString{String: emp.UpdatedBy.String(), Valid: true}
	}

	return row
}

// Create cria um novo funcionário
func (repo *EmployeeRepository) Create(ctx context.Context, emp *employee.Employee) error {
	row := repo.fromEntity(emp)

	query := `
		INSERT INTO employees (
			id, tenant_id, full_name, identity, identity_type,
			date_of_birth, photo_url, face_embedding, phone, email,
			active, created_at, updated_at, created_by, updated_by
		) VALUES (
			:id, :tenant_id, :full_name, :identity, :identity_type,
			:date_of_birth, :photo_url, :face_embedding, :phone, :email,
			:active, :created_at, :updated_at, :created_by, :updated_by
		)`

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create employee", zap.Error(err), zap.String("employee_id", emp.ID.String()))
		return errors.NewInternalError("failed to create employee", err)
	}

	repo.logger.Info("Employee created successfully", zap.String("employee_id", emp.ID.String()))
	return nil
}

// GetByID busca um funcionário pelo ID
func (repo *EmployeeRepository) GetByID(ctx context.Context, id value_objects.UUID) (*employee.Employee, error) {
	var row employeeRow

	query := `
		SELECT id, tenant_id, full_name, identity, identity_type,
			   date_of_birth, photo_url, face_embedding, phone, email,
			   active, created_at, updated_at, created_by, updated_by
		FROM employees 
		WHERE id = $1 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "employee not found", nil)
		}
		repo.logger.Error("Failed to get employee by ID", zap.Error(err), zap.String("employee_id", id.String()))
		return nil, errors.NewInternalError("failed to get employee", err)
	}

	return row.toEntity()
}

// GetByIDAndTenant busca um funcionário pelo ID dentro de um tenant
func (repo *EmployeeRepository) GetByIDAndTenant(ctx context.Context, id, tenantID value_objects.UUID) (*employee.Employee, error) {
	var row employeeRow

	query := `
		SELECT id, tenant_id, full_name, identity, identity_type,
			   date_of_birth, photo_url, face_embedding, phone, email,
			   active, created_at, updated_at, created_by, updated_by
		FROM employees 
		WHERE id = $1 AND tenant_id = $2 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, id.String(), tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "employee not found", nil)
		}
		repo.logger.Error("Failed to get employee by ID and tenant",
			zap.Error(err),
			zap.String("employee_id", id.String()),
			zap.String("tenant_id", tenantID.String()))
		return nil, errors.NewInternalError("failed to get employee", err)
	}

	return row.toEntity()
}

// GetByIdentity busca um funcionário pela identidade
func (repo *EmployeeRepository) GetByIdentity(ctx context.Context, identity string) (*employee.Employee, error) {
	var row employeeRow

	query := `
		SELECT id, tenant_id, full_name, identity, identity_type,
			   date_of_birth, photo_url, face_embedding, phone, email,
			   active, created_at, updated_at, created_by, updated_by
		FROM employees 
		WHERE identity = $1 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, identity)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "employee not found", nil)
		}
		repo.logger.Error("Failed to get employee by identity", zap.Error(err), zap.String("identity", identity))
		return nil, errors.NewInternalError("failed to get employee", err)
	}

	return row.toEntity()
}

// GetByIdentityAndTenant busca um funcionário pela identidade dentro de um tenant
func (repo *EmployeeRepository) GetByIdentityAndTenant(ctx context.Context, identity string, tenantID value_objects.UUID) (*employee.Employee, error) {
	var row employeeRow

	query := `
		SELECT id, tenant_id, full_name, identity, identity_type,
			   date_of_birth, photo_url, face_embedding, phone, email,
			   active, created_at, updated_at, created_by, updated_by
		FROM employees 
		WHERE identity = $1 AND tenant_id = $2 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, identity, tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "employee not found", nil)
		}
		repo.logger.Error("Failed to get employee by identity and tenant",
			zap.Error(err),
			zap.String("identity", identity),
			zap.String("tenant_id", tenantID.String()))
		return nil, errors.NewInternalError("failed to get employee", err)
	}

	return row.toEntity()
}

// GetByEmail busca um funcionário pelo email
func (repo *EmployeeRepository) GetByEmail(ctx context.Context, email string) (*employee.Employee, error) {
	var row employeeRow

	query := `
		SELECT id, tenant_id, full_name, identity, identity_type,
			   date_of_birth, photo_url, face_embedding, phone, email,
			   active, created_at, updated_at, created_by, updated_by
		FROM employees 
		WHERE email = $1 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "employee not found", nil)
		}
		repo.logger.Error("Failed to get employee by email", zap.Error(err), zap.String("email", email))
		return nil, errors.NewInternalError("failed to get employee", err)
	}

	return row.toEntity()
}

// GetByEmailAndTenant busca um funcionário pelo email dentro de um tenant
func (repo *EmployeeRepository) GetByEmailAndTenant(ctx context.Context, email string, tenantID value_objects.UUID) (*employee.Employee, error) {
	var row employeeRow

	query := `
		SELECT id, tenant_id, full_name, identity, identity_type,
			   date_of_birth, photo_url, face_embedding, phone, email,
			   active, created_at, updated_at, created_by, updated_by
		FROM employees 
		WHERE email = $1 AND tenant_id = $2 AND active = true`

	err := repo.db.GetContext(ctx, &row, query, email, tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewDomainError("NOT_FOUND", "employee not found", nil)
		}
		repo.logger.Error("Failed to get employee by email and tenant",
			zap.Error(err),
			zap.String("email", email),
			zap.String("tenant_id", tenantID.String()))
		return nil, errors.NewInternalError("failed to get employee", err)
	}

	return row.toEntity()
}

// Update atualiza um funcionário existente
func (repo *EmployeeRepository) Update(ctx context.Context, emp *employee.Employee) error {
	row := repo.fromEntity(emp)

	query := `
		UPDATE employees SET
			full_name = :full_name,
			identity = :identity,
			identity_type = :identity_type,
			date_of_birth = :date_of_birth,
			photo_url = :photo_url,
			face_embedding = :face_embedding,
			phone = :phone,
			email = :email,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id = :id AND active = true`

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update employee", zap.Error(err), zap.String("employee_id", emp.ID.String()))
		return errors.NewInternalError("failed to update employee", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repo.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError("failed to update employee", err)
	}

	if rowsAffected == 0 {
		return errors.NewDomainError("NOT_FOUND", "employee not found or inactive", nil)
	}

	repo.logger.Info("Employee updated successfully", zap.String("employee_id", emp.ID.String()))
	return nil
}

// Delete remove um funcionário (soft delete)
func (repo *EmployeeRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE employees SET
			active = false,
			updated_at = NOW(),
			updated_by = $2
		WHERE id = $1 AND active = true`

	result, err := repo.db.ExecContext(ctx, query, id.String(), deletedBy.String())
	if err != nil {
		repo.logger.Error("Failed to delete employee", zap.Error(err), zap.String("employee_id", id.String()))
		return errors.NewInternalError("failed to delete employee", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		repo.logger.Error("Failed to get rows affected", zap.Error(err))
		return errors.NewInternalError("failed to delete employee", err)
	}

	if rowsAffected == 0 {
		return errors.NewDomainError("NOT_FOUND", "employee not found or already inactive", nil)
	}

	repo.logger.Info("Employee deleted successfully", zap.String("employee_id", id.String()))
	return nil
}

// List lista funcionários com paginação e filtros
func (repo *EmployeeRepository) List(ctx context.Context, filters employee.ListFilters) ([]*employee.Employee, int, error) {
	if err := filters.Validate(); err != nil {
		return nil, 0, err
	}

	// Query base
	baseQuery := `FROM employees WHERE active = true`
	var args []interface{}
	var conditions []string
	argIndex := 1

	// Aplicar filtros
	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
		args = append(args, filters.TenantID.String())
		argIndex++
	}

	if filters.FullName != nil {
		conditions = append(conditions, fmt.Sprintf("full_name ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.FullName+"%")
		argIndex++
	}

	if filters.Identity != nil {
		conditions = append(conditions, fmt.Sprintf("identity = $%d", argIndex))
		args = append(args, *filters.Identity)
		argIndex++
	}

	if filters.IdentityType != nil {
		conditions = append(conditions, fmt.Sprintf("identity_type = $%d", argIndex))
		args = append(args, *filters.IdentityType)
		argIndex++
	}

	if filters.Email != nil {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Email+"%")
		argIndex++
	}

	if filters.Phone != nil {
		conditions = append(conditions, fmt.Sprintf("phone ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Phone+"%")
		argIndex++
	}

	if filters.HasPhoto != nil {
		if *filters.HasPhoto {
			conditions = append(conditions, "photo_url IS NOT NULL AND photo_url != ''")
		} else {
			conditions = append(conditions, "(photo_url IS NULL OR photo_url = '')")
		}
	}

	if filters.HasFaceEmbedding != nil {
		if *filters.HasFaceEmbedding {
			conditions = append(conditions, "face_embedding IS NOT NULL AND array_length(face_embedding, 1) > 0")
		} else {
			conditions = append(conditions, "(face_embedding IS NULL OR array_length(face_embedding, 1) = 0)")
		}
	}

	if filters.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filters.Active)
		argIndex++
	}

	// Construir WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Query de contagem
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	var total int
	err := repo.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to count employees", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to count employees", err)
	}

	// Query de dados
	orderClause := fmt.Sprintf("ORDER BY %s", filters.OrderBy)
	if filters.OrderDesc {
		orderClause += " DESC"
	}

	limitClause := fmt.Sprintf("LIMIT %d OFFSET %d", filters.PageSize, filters.GetOffset())

	dataQuery := `
		SELECT id, tenant_id, full_name, identity, identity_type,
			   date_of_birth, photo_url, face_embedding, phone, email,
			   active, created_at, updated_at, created_by, updated_by ` +
		baseQuery + whereClause + " " + orderClause + " " + limitClause

	var rows []employeeRow
	err = repo.db.SelectContext(ctx, &rows, dataQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list employees", zap.Error(err))
		return nil, 0, errors.NewInternalError("failed to list employees", err)
	}

	// Converter para entidades
	employees := make([]*employee.Employee, 0, len(rows))
	for _, row := range rows {
		emp, err := row.toEntity()
		if err != nil {
			repo.logger.Warn("Failed to convert employee row", zap.Error(err), zap.String("employee_id", row.ID))
			continue
		}
		employees = append(employees, emp)
	}

	return employees, total, nil
}

// ListByTenant lista funcionários de um tenant específico
func (repo *EmployeeRepository) ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters employee.ListFilters) ([]*employee.Employee, int, error) {
	filters.TenantID = &tenantID
	return repo.List(ctx, filters)
}

// ListByPartner lista funcionários de um parceiro específico
func (repo *EmployeeRepository) ListByPartner(ctx context.Context, partnerID value_objects.UUID, filters employee.ListFilters) ([]*employee.Employee, int, error) {
	// TODO: Implementar quando houver relacionamento Employee-Partner
	// Por enquanto, retorna lista vazia
	return []*employee.Employee{}, 0, nil
}

// ListByEvent lista funcionários associados a um evento (através de parceiros)
func (repo *EmployeeRepository) ListByEvent(ctx context.Context, eventID value_objects.UUID, filters employee.ListFilters) ([]*employee.Employee, int, error) {
	// TODO: Implementar quando houver relacionamento Employee-Event
	// Por enquanto, retorna lista vazia
	return []*employee.Employee{}, 0, nil
}

// ExistsByIdentity verifica se existe um funcionário com a identidade informada
func (repo *EmployeeRepository) ExistsByIdentity(ctx context.Context, identity string, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM employees WHERE identity = $1 AND active = true`
	args := []interface{}{identity}

	if excludeID != nil {
		query += " AND id != $2"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check employee identity existence", zap.Error(err))
		return false, errors.NewInternalError("failed to check employee identity", err)
	}

	return count > 0, nil
}

// ExistsByIdentityInTenant verifica se existe um funcionário com a identidade no tenant
func (repo *EmployeeRepository) ExistsByIdentityInTenant(ctx context.Context, identity string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM employees WHERE identity = $1 AND tenant_id = $2 AND active = true`
	args := []interface{}{identity, tenantID.String()}

	if excludeID != nil {
		query += " AND id != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check employee identity existence in tenant", zap.Error(err))
		return false, errors.NewInternalError("failed to check employee identity", err)
	}

	return count > 0, nil
}

// ExistsByEmail verifica se existe um funcionário com o email informado
func (repo *EmployeeRepository) ExistsByEmail(ctx context.Context, email string, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM employees WHERE email = $1 AND active = true`
	args := []interface{}{email}

	if excludeID != nil {
		query += " AND id != $2"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check employee email existence", zap.Error(err))
		return false, errors.NewInternalError("failed to check employee email", err)
	}

	return count > 0, nil
}

// ExistsByEmailInTenant verifica se existe um funcionário com o email no tenant
func (repo *EmployeeRepository) ExistsByEmailInTenant(ctx context.Context, email string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM employees WHERE email = $1 AND tenant_id = $2 AND active = true`
	args := []interface{}{email, tenantID.String()}

	if excludeID != nil {
		query += " AND id != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		repo.logger.Error("Failed to check employee email existence in tenant", zap.Error(err))
		return false, errors.NewInternalError("failed to check employee email", err)
	}

	return count > 0, nil
}

// FindByFaceEmbedding busca funcionários similares por embedding facial
func (repo *EmployeeRepository) FindByFaceEmbedding(ctx context.Context, embedding []float32, tenantID *value_objects.UUID, threshold float32, limit int) ([]*employee.Employee, []float32, error) {
	// TODO: Implementar busca por similaridade facial usando PostGIS ou extensão de vetores
	// Por enquanto, retorna lista vazia
	return []*employee.Employee{}, []float32{}, nil
}

// GetEmployeesWithFaceEmbedding busca funcionários que têm embedding facial
func (repo *EmployeeRepository) GetEmployeesWithFaceEmbedding(ctx context.Context, tenantID *value_objects.UUID, filters employee.ListFilters) ([]*employee.Employee, int, error) {
	hasFaceEmbedding := true
	filters.HasFaceEmbedding = &hasFaceEmbedding
	if tenantID != nil {
		filters.TenantID = tenantID
	}
	return repo.List(ctx, filters)
}
