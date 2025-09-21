package employee

import (
	"time"

	"eventos-backend/internal/domain/shared/constants"
	"eventos-backend/internal/domain/shared/errors"
	"eventos-backend/internal/domain/shared/value_objects"
)

// Employee representa um funcionário no sistema
type Employee struct {
	ID            value_objects.UUID
	TenantID      value_objects.UUID
	FullName      string
	Identity      string
	IdentityType  string
	DateOfBirth   *time.Time
	PhotoURL      string
	FaceEmbedding []float32 // Embedding facial para reconhecimento (512 dimensões)
	Phone         string
	Email         string
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CreatedBy     *value_objects.UUID
	UpdatedBy     *value_objects.UUID
}

// NewEmployee cria uma nova instância de Employee
func NewEmployee(tenantID value_objects.UUID, fullName, identity, identityType, phone, email string, dateOfBirth *time.Time, createdBy value_objects.UUID) (*Employee, error) {
	if err := validateEmployeeData(fullName, identity, identityType, phone, email, dateOfBirth); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	return &Employee{
		ID:           value_objects.NewUUID(),
		TenantID:     tenantID,
		FullName:     fullName,
		Identity:     identity,
		IdentityType: identityType,
		DateOfBirth:  dateOfBirth,
		Phone:        phone,
		Email:        email,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    &createdBy,
		UpdatedBy:    &createdBy,
	}, nil
}

// Update atualiza os dados do funcionário
func (e *Employee) Update(fullName, identity, identityType, phone, email string, dateOfBirth *time.Time, updatedBy value_objects.UUID) error {
	if err := validateEmployeeData(fullName, identity, identityType, phone, email, dateOfBirth); err != nil {
		return err
	}

	e.FullName = fullName
	e.Identity = identity
	e.IdentityType = identityType
	e.Phone = phone
	e.Email = email
	e.DateOfBirth = dateOfBirth
	e.UpdatedAt = time.Now().UTC()
	e.UpdatedBy = &updatedBy

	return nil
}

// UpdatePhoto atualiza a foto do funcionário
func (e *Employee) UpdatePhoto(photoURL string, updatedBy value_objects.UUID) error {
	if photoURL != "" && len(photoURL) > 500 {
		return errors.NewValidationError("photo_url", "photo URL must be at most 500 characters")
	}

	e.PhotoURL = photoURL
	e.UpdatedAt = time.Now().UTC()
	e.UpdatedBy = &updatedBy

	return nil
}

// UpdateFaceEmbedding atualiza o embedding facial do funcionário
func (e *Employee) UpdateFaceEmbedding(embedding []float32, updatedBy value_objects.UUID) error {
	if len(embedding) != 512 {
		return errors.NewValidationError("face_embedding", "face embedding must have exactly 512 dimensions")
	}

	// Validar que todos os valores são números válidos
	for i, val := range embedding {
		if val != val { // Verificar NaN
			return errors.NewValidationError("face_embedding", "face embedding contains invalid values")
		}
		if val < -1.0 || val > 1.0 {
			return errors.NewValidationError("face_embedding", "face embedding values must be between -1.0 and 1.0")
		}
		// Limitar precisão para evitar problemas de armazenamento
		embedding[i] = float32(int(val*10000)) / 10000
	}

	e.FaceEmbedding = embedding
	e.UpdatedAt = time.Now().UTC()
	e.UpdatedBy = &updatedBy

	return nil
}

// Activate ativa o funcionário
func (e *Employee) Activate(updatedBy value_objects.UUID) {
	e.Active = true
	e.UpdatedAt = time.Now().UTC()
	e.UpdatedBy = &updatedBy
}

// Deactivate desativa o funcionário
func (e *Employee) Deactivate(updatedBy value_objects.UUID) {
	e.Active = false
	e.UpdatedAt = time.Now().UTC()
	e.UpdatedBy = &updatedBy
}

// IsActive verifica se o funcionário está ativo
func (e *Employee) IsActive() bool {
	return e.Active
}

// BelongsToTenant verifica se o funcionário pertence ao tenant informado
func (e *Employee) BelongsToTenant(tenantID value_objects.UUID) bool {
	return e.TenantID.Equals(tenantID)
}

// HasPhoto verifica se o funcionário tem foto
func (e *Employee) HasPhoto() bool {
	return e.PhotoURL != ""
}

// HasFaceEmbedding verifica se o funcionário tem embedding facial
func (e *Employee) HasFaceEmbedding() bool {
	return len(e.FaceEmbedding) == 512
}

// GetAge calcula a idade do funcionário
func (e *Employee) GetAge() int {
	if e.DateOfBirth == nil {
		return 0
	}

	now := time.Now()
	age := now.Year() - e.DateOfBirth.Year()

	// Ajustar se ainda não fez aniversário este ano
	if now.YearDay() < e.DateOfBirth.YearDay() {
		age--
	}

	return age
}

// IsMinor verifica se o funcionário é menor de idade
func (e *Employee) IsMinor() bool {
	return e.GetAge() < 18
}

// CanPerformFacialRecognition verifica se o funcionário pode usar reconhecimento facial
func (e *Employee) CanPerformFacialRecognition() bool {
	return e.IsActive() && e.HasFaceEmbedding()
}

// CompareFaceEmbedding compara o embedding facial com outro embedding
func (e *Employee) CompareFaceEmbedding(otherEmbedding []float32, threshold float32) (bool, float32) {
	if !e.HasFaceEmbedding() || len(otherEmbedding) != 512 {
		return false, 0.0
	}

	// Calcular similaridade coseno
	similarity := cosineSimilarity(e.FaceEmbedding, otherEmbedding)

	return similarity >= threshold, similarity
}

// validateEmployeeData valida os dados básicos do funcionário
func validateEmployeeData(fullName, identity, identityType, phone, email string, dateOfBirth *time.Time) error {
	if fullName == "" {
		return errors.NewValidationError("full_name", "full name is required")
	}

	if len(fullName) < 2 || len(fullName) > 255 {
		return errors.NewValidationError("full_name", "full name must be between 2 and 255 characters")
	}

	if identity != "" {
		if len(identity) < 3 || len(identity) > 50 {
			return errors.NewValidationError("identity", "identity must be between 3 and 50 characters")
		}

		if !isValidIdentityType(identityType) {
			return errors.NewValidationError("identity_type", "invalid identity type")
		}
	}

	if phone != "" {
		if len(phone) < 8 || len(phone) > 20 {
			return errors.NewValidationError("phone", "phone must be between 8 and 20 characters")
		}
	}

	if email != "" {
		if !isValidEmail(email) {
			return errors.NewValidationError("email", "invalid email format")
		}
	}

	if dateOfBirth != nil {
		// Verificar se a data de nascimento não é futura
		if dateOfBirth.After(time.Now()) {
			return errors.NewValidationError("date_of_birth", "date of birth cannot be in the future")
		}

		// Verificar se a pessoa não é muito velha (máximo 120 anos)
		if time.Since(*dateOfBirth) > 120*365*24*time.Hour {
			return errors.NewValidationError("date_of_birth", "date of birth is too old")
		}

		// Verificar idade mínima (14 anos para trabalhar)
		age := time.Now().Year() - dateOfBirth.Year()
		if age < 14 {
			return errors.NewValidationError("date_of_birth", "employee must be at least 14 years old")
		}
	}

	return nil
}

// isValidEmail faz uma validação básica de email
func isValidEmail(email string) bool {
	if len(email) < 5 || len(email) > 255 {
		return false
	}

	atCount := 0
	atIndex := -1
	for i, char := range email {
		if char == '@' {
			atCount++
			atIndex = i
		}
	}

	if atCount != 1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	// Deve ter pelo menos um ponto após o @
	hasDotAfterAt := false
	for i := atIndex + 1; i < len(email); i++ {
		if email[i] == '.' && i < len(email)-1 {
			hasDotAfterAt = true
			break
		}
	}

	return hasDotAfterAt
}

// isValidIdentityType verifica se o tipo de identidade é válido
func isValidIdentityType(identityType string) bool {
	validTypes := []string{
		constants.IdentityTypeCPF,
		constants.IdentityTypeCNPJ,
		constants.IdentityTypeRG,
		constants.IdentityTypeOther,
	}

	for _, validType := range validTypes {
		if identityType == validType {
			return true
		}
	}

	return false
}

// cosineSimilarity calcula a similaridade coseno entre dois vetores
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (sqrt32(normA) * sqrt32(normB))
}

// sqrt32 calcula a raiz quadrada de um float32
func sqrt32(x float32) float32 {
	if x == 0 {
		return 0
	}

	// Método de Newton para float32
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
