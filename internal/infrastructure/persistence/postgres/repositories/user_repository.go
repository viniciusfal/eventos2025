package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"eventos-backend/internal/domain/shared/value_objects"
	"eventos-backend/internal/domain/user"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// UserRepository implementa a interface de repositório para User
type UserRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewUserRepository cria uma nova instância do repositório de user
func NewUserRepository(db *sqlx.DB, logger *zap.Logger) user.Repository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// userRow representa uma linha da tabela user no banco
type userRow struct {
	ID        string         `db:"id_user"`
	TenantID  string         `db:"id_tenant"`
	FullName  string         `db:"full_name"`
	Email     string         `db:"email"`
	Phone     sql.NullString `db:"phone"`
	Username  string         `db:"username"`
	Password  string         `db:"password"`
	Active    bool           `db:"active"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	CreatedBy sql.NullString `db:"created_by"`
	UpdatedBy sql.NullString `db:"updated_by"`
}

// toEntity converte uma linha do banco para entidade de domínio
func (r *userRow) toEntity() (*user.User, error) {
	id, err := value_objects.ParseUUID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	tenantID, err := value_objects.ParseUUID(r.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	u := &user.User{
		ID:        id,
		TenantID:  tenantID,
		FullName:  r.FullName,
		Email:     r.Email,
		Username:  r.Username,
		Password:  r.Password,
		Active:    r.Active,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}

	if r.Phone.Valid {
		u.Phone = r.Phone.String
	}

	if r.CreatedBy.Valid {
		createdBy, err := value_objects.ParseUUID(r.CreatedBy.String)
		if err == nil {
			u.CreatedBy = &createdBy
		}
	}

	if r.UpdatedBy.Valid {
		updatedBy, err := value_objects.ParseUUID(r.UpdatedBy.String)
		if err == nil {
			u.UpdatedBy = &updatedBy
		}
	}

	return u, nil
}

// fromEntity converte uma entidade de domínio para linha do banco
func (repo *UserRepository) fromEntity(u *user.User) *userRow {
	row := &userRow{
		ID:        u.ID.String(),
		TenantID:  u.TenantID.String(),
		FullName:  u.FullName,
		Email:     u.Email,
		Username:  u.Username,
		Password:  u.Password,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}

	if u.Phone != "" {
		row.Phone = sql.NullString{String: u.Phone, Valid: true}
	}

	if u.CreatedBy != nil && !u.CreatedBy.IsZero() {
		row.CreatedBy = sql.NullString{String: u.CreatedBy.String(), Valid: true}
	}

	if u.UpdatedBy != nil && !u.UpdatedBy.IsZero() {
		row.UpdatedBy = sql.NullString{String: u.UpdatedBy.String(), Valid: true}
	}

	return row
}

// Create cria um novo usuário
func (repo *UserRepository) Create(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO user_account (
			id_user, id_tenant, full_name, email, phone, username, 
			password, active, created_at, updated_at, created_by, updated_by
		) VALUES (
			:id_user, :id_tenant, :full_name, :email, :phone, :username,
			:password, :active, :created_at, :updated_at, :created_by, :updated_by
		)`

	row := repo.fromEntity(u)

	_, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to create user", zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID busca um usuário pelo ID
func (repo *UserRepository) GetByID(ctx context.Context, id value_objects.UUID) (*user.User, error) {
	query := `
		SELECT id_user, id_tenant, full_name, email, phone, username,
		       password, active, created_at, updated_at, created_by, updated_by
		FROM user_account 
		WHERE id_user = $1 AND active = true`

	var row userRow
	err := repo.db.GetContext(ctx, &row, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		repo.logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return row.toEntity()
}

// GetByUsername busca um usuário pelo username
func (repo *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	query := `
		SELECT id_user, id_tenant, full_name, email, phone, username,
		       password, active, created_at, updated_at, created_by, updated_by
		FROM user_account 
		WHERE username = $1 AND active = true`

	var row userRow
	err := repo.db.GetContext(ctx, &row, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		repo.logger.Error("Failed to get user by username", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return row.toEntity()
}

// GetByEmail busca um usuário pelo email
func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT id_user, id_tenant, full_name, email, phone, username,
		       password, active, created_at, updated_at, created_by, updated_by
		FROM user_account 
		WHERE email = $1 AND active = true`

	var row userRow
	err := repo.db.GetContext(ctx, &row, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		repo.logger.Error("Failed to get user by email", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return row.toEntity()
}

// GetByUsernameAndTenant busca um usuário pelo username dentro de um tenant
func (repo *UserRepository) GetByUsernameAndTenant(ctx context.Context, username string, tenantID value_objects.UUID) (*user.User, error) {
	query := `
		SELECT id_user, id_tenant, full_name, email, phone, username,
		       password, active, created_at, updated_at, created_by, updated_by
		FROM user_account 
		WHERE username = $1 AND id_tenant = $2 AND active = true`

	var row userRow
	err := repo.db.GetContext(ctx, &row, query, username, tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		repo.logger.Error("Failed to get user by username and tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return row.toEntity()
}

// GetByEmailAndTenant busca um usuário pelo email dentro de um tenant
func (repo *UserRepository) GetByEmailAndTenant(ctx context.Context, email string, tenantID value_objects.UUID) (*user.User, error) {
	query := `
		SELECT id_user, id_tenant, full_name, email, phone, username,
		       password, active, created_at, updated_at, created_by, updated_by
		FROM user_account 
		WHERE email = $1 AND id_tenant = $2 AND active = true`

	var row userRow
	err := repo.db.GetContext(ctx, &row, query, email, tenantID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		repo.logger.Error("Failed to get user by email and tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return row.toEntity()
}

// Update atualiza um usuário existente
func (repo *UserRepository) Update(ctx context.Context, u *user.User) error {
	query := `
		UPDATE user_account SET
			full_name = :full_name,
			email = :email,
			phone = :phone,
			username = :username,
			password = :password,
			active = :active,
			updated_at = :updated_at,
			updated_by = :updated_by
		WHERE id_user = :id_user`

	row := repo.fromEntity(u)

	result, err := repo.db.NamedExecContext(ctx, query, row)
	if err != nil {
		repo.logger.Error("Failed to update user", zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete remove um usuário (soft delete)
func (repo *UserRepository) Delete(ctx context.Context, id value_objects.UUID, deletedBy value_objects.UUID) error {
	query := `
		UPDATE user_account SET
			active = false,
			updated_at = $1,
			updated_by = $2
		WHERE id_user = $3`

	result, err := repo.db.ExecContext(ctx, query, time.Now().UTC(), deletedBy.String(), id.String())
	if err != nil {
		repo.logger.Error("Failed to delete user", zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List lista usuários com paginação e filtros
func (repo *UserRepository) List(ctx context.Context, filters user.ListFilters) ([]*user.User, int, error) {
	// Construir query base
	baseQuery := `
		SELECT id_user, id_tenant, full_name, email, phone, username,
		       password, active, created_at, updated_at, created_by, updated_by
		FROM user_account`

	countQuery := "SELECT COUNT(*) FROM user_account"

	// Construir condições WHERE
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf("id_tenant = $%d", argIndex))
		args = append(args, filters.TenantID.String())
		argIndex++
	}

	if filters.FullName != nil {
		conditions = append(conditions, fmt.Sprintf("full_name ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.FullName+"%")
		argIndex++
	}

	if filters.Email != nil {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Email+"%")
		argIndex++
	}

	if filters.Username != nil {
		conditions = append(conditions, fmt.Sprintf("username ILIKE $%d", argIndex))
		args = append(args, "%"+*filters.Username+"%")
		argIndex++
	}

	if filters.Active != nil {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *filters.Active)
		argIndex++
	}

	// Adicionar WHERE se houver condições
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Contar total de registros
	var total int
	err := repo.db.GetContext(ctx, &total, countQuery+whereClause, args...)
	if err != nil {
		repo.logger.Error("Failed to count users", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Construir query com ordenação e paginação
	orderClause := fmt.Sprintf(" ORDER BY %s", filters.OrderBy)
	if filters.OrderDesc {
		orderClause += " DESC"
	}

	limitClause := fmt.Sprintf(" LIMIT %d OFFSET %d", filters.PageSize, filters.GetOffset())

	finalQuery := baseQuery + whereClause + orderClause + limitClause

	// Executar query
	var rows []userRow
	err = repo.db.SelectContext(ctx, &rows, finalQuery, args...)
	if err != nil {
		repo.logger.Error("Failed to list users", zap.Error(err))
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Converter para entidades
	var users []*user.User
	for _, row := range rows {
		u, err := row.toEntity()
		if err != nil {
			repo.logger.Error("Failed to convert user row to entity", zap.Error(err))
			continue
		}
		users = append(users, u)
	}

	return users, total, nil
}

// Implementações básicas para satisfazer a interface (podem ser implementadas depois)
func (repo *UserRepository) ExistsByUsername(ctx context.Context, username string, excludeID *value_objects.UUID) (bool, error) {
	query := "SELECT COUNT(*) FROM user_account WHERE username = $1 AND active = true"
	args := []interface{}{username}

	if excludeID != nil {
		query += " AND id_user != $2"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	return count > 0, err
}

func (repo *UserRepository) ExistsByEmail(ctx context.Context, email string, excludeID *value_objects.UUID) (bool, error) {
	query := "SELECT COUNT(*) FROM user_account WHERE email = $1 AND active = true"
	args := []interface{}{email}

	if excludeID != nil {
		query += " AND id_user != $2"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	return count > 0, err
}

func (repo *UserRepository) ExistsByUsernameInTenant(ctx context.Context, username string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error) {
	query := "SELECT COUNT(*) FROM user_account WHERE username = $1 AND id_tenant = $2 AND active = true"
	args := []interface{}{username, tenantID.String()}

	if excludeID != nil {
		query += " AND id_user != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	return count > 0, err
}

func (repo *UserRepository) ExistsByEmailInTenant(ctx context.Context, email string, tenantID value_objects.UUID, excludeID *value_objects.UUID) (bool, error) {
	query := "SELECT COUNT(*) FROM user_account WHERE email = $1 AND id_tenant = $2 AND active = true"
	args := []interface{}{email, tenantID.String()}

	if excludeID != nil {
		query += " AND id_user != $3"
		args = append(args, excludeID.String())
	}

	var count int
	err := repo.db.GetContext(ctx, &count, query, args...)
	return count > 0, err
}

func (repo *UserRepository) ListByTenant(ctx context.Context, tenantID value_objects.UUID, filters user.ListFilters) ([]*user.User, int, error) {
	filters.TenantID = &tenantID
	return repo.List(ctx, filters)
}
