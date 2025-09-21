-- Migration: 001_create_database_schema.sql
-- Database: PostgreSQL
-- Description: Criação inicial do schema do banco de dados para o sistema de check-in em eventos

-- Criando extensões necessárias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- Tabela de módulos do sistema
CREATE TABLE module (
    id_module UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name_module VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    active BOOLEAN DEFAULT true
);

-- Tabela de tenants (organizações)
CREATE TABLE tenant (
    id_tenant UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_config_tenant UUID,
    name VARCHAR(255) NOT NULL,
    identity VARCHAR(50),
    type_identity VARCHAR(20),
    email VARCHAR(255),
    address TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de configurações dos tenants
CREATE TABLE config_tenant (
    id_config_tenant UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    modules JSONB,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by UUID,
    FOREIGN KEY (updated_by) REFERENCES tenant(id_tenant)
);

-- Adicionando a FK que estava faltando na tabela tenant
ALTER TABLE tenant ADD FOREIGN KEY (id_config_tenant) REFERENCES config_tenant(id_config_tenant);

-- Tabela de usuários do sistema
CREATE TABLE "user" (
    id_user UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20),
    username VARCHAR(50) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    FOREIGN KEY (created_by) REFERENCES "user"(id_user),
    FOREIGN KEY (updated_by) REFERENCES "user"(id_user)
);

-- Tabela de papéis (roles)
CREATE TABLE role (
    id_role UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    UNIQUE(id_tenant, name)
);

-- Tabela de permissões
CREATE TABLE permission (
    id_permission UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    module VARCHAR(100)
);

-- Tabela de relacionamento entre papéis e permissões
CREATE TABLE role_permission (
    id_role UUID,
    id_permission UUID,
    PRIMARY KEY (id_role, id_permission),
    FOREIGN KEY (id_role) REFERENCES role(id_role) ON DELETE CASCADE,
    FOREIGN KEY (id_permission) REFERENCES permission(id_permission) ON DELETE CASCADE
);

-- Tabela de relacionamento entre usuários e papéis
CREATE TABLE user_role (
    id_user UUID,
    id_role UUID,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by UUID,
    PRIMARY KEY (id_user, id_role),
    FOREIGN KEY (id_user) REFERENCES "user"(id_user) ON DELETE CASCADE,
    FOREIGN KEY (id_role) REFERENCES role(id_role) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES "user"(id_user)
);

-- Tabela de eventos
CREATE TABLE event (
    id_event UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    location TEXT,
    fence_event GEOGRAPHY(POLYGON, 4326),
    initial_date TIMESTAMP NOT NULL,
    final_date TIMESTAMP NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    FOREIGN KEY (created_by) REFERENCES "user"(id_user),
    FOREIGN KEY (updated_by) REFERENCES "user"(id_user)
);

-- Tabela de parceiros
CREATE TABLE partner (
    id_partner UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID NOT NULL,
    name_partner VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    email_2 VARCHAR(255),
    phone VARCHAR(20),
    phone_2 VARCHAR(20),
    identity VARCHAR(50),
    type_identity VARCHAR(20),
    location TEXT,
    pass_hash VARCHAR(255),
    last_login TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    FOREIGN KEY (created_by) REFERENCES "user"(id_user),
    FOREIGN KEY (updated_by) REFERENCES "user"(id_user)
);

-- Tabela de funcionários
CREATE TABLE employee (
    id_employee UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    identity VARCHAR(50),
    type_identity VARCHAR(20),
    date_of_birth DATE,
    photo_url TEXT,
    face_embedding VECTOR(512), -- Requer extensão pgvector
    phone VARCHAR(20),
    email VARCHAR(255),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    updated_by UUID,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    FOREIGN KEY (created_by) REFERENCES "user"(id_user),
    FOREIGN KEY (updated_by) REFERENCES "user"(id_user)
);

-- Tabela de relacionamento entre eventos e parceiros
CREATE TABLE event_partner (
    id_event UUID,
    id_partner UUID,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by UUID,
    PRIMARY KEY (id_event, id_partner),
    FOREIGN KEY (id_event) REFERENCES event(id_event) ON DELETE CASCADE,
    FOREIGN KEY (id_partner) REFERENCES partner(id_partner) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES "user"(id_user)
);

-- Tabela de relacionamento entre parceiros e funcionários
CREATE TABLE partner_employee (
    id_partner UUID,
    id_employee UUID,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by UUID,
    PRIMARY KEY (id_partner, id_employee),
    FOREIGN KEY (id_partner) REFERENCES partner(id_partner) ON DELETE CASCADE,
    FOREIGN KEY (id_employee) REFERENCES employee(id_employee) ON DELETE CASCADE,
    FOREIGN KEY (assigned_by) REFERENCES "user"(id_user)
);

-- Tabela de QR Codes para eventos
CREATE TABLE event_qr_code (
    id_qr_code UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_event UUID NOT NULL,
    qr_type VARCHAR(10) NOT NULL, -- "checkin" ou "checkout"
    qr_token VARCHAR(255) NOT NULL UNIQUE,
    valid_from TIMESTAMP NOT NULL,
    valid_until TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by UUID,
    FOREIGN KEY (id_event) REFERENCES event(id_event),
    FOREIGN KEY (created_by) REFERENCES "user"(id_user)
);

-- Tabela de check-ins
CREATE TABLE checkin (
    id_checkin UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID NOT NULL,
    id_event UUID NOT NULL,
    id_employee UUID NOT NULL,
    id_partner UUID NOT NULL,
    user_resp_checkin UUID,
    checkin_date_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    location GEOGRAPHY(POINT, 4326),
    device_id VARCHAR(100),
    notes TEXT,
    check_method VARCHAR(20) DEFAULT 'facial_recognition',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    FOREIGN KEY (id_event) REFERENCES event(id_event),
    FOREIGN KEY (id_employee) REFERENCES employee(id_employee),
    FOREIGN KEY (id_partner) REFERENCES partner(id_partner),
    FOREIGN KEY (user_resp_checkin) REFERENCES "user"(id_user)
);

-- Tabela de check-outs
CREATE TABLE checkout (
    id_checkout UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID NOT NULL,
    id_event UUID NOT NULL,
    id_employee UUID NOT NULL,
    id_partner UUID NOT NULL,
    user_resp_checkout UUID,
    checkout_date_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    location_checkout GEOGRAPHY(POINT, 4326),
    device_id VARCHAR(100),
    notes TEXT,
    check_method VARCHAR(20) DEFAULT 'facial_recognition',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    FOREIGN KEY (id_event) REFERENCES event(id_event),
    FOREIGN KEY (id_employee) REFERENCES employee(id_employee),
    FOREIGN KEY (id_partner) REFERENCES partner(id_partner),
    FOREIGN KEY (user_resp_checkout) REFERENCES "user"(id_user)
);

-- Tabela de logs de eventos
CREATE TABLE event_log (
    id_event_log UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID NOT NULL,
    id_event UUID NOT NULL,
    log_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    details JSONB,
    user_id UUID,
    ip_address VARCHAR(45),
    user_agent TEXT,
    location GEOGRAPHY(POINT, 4326),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    FOREIGN KEY (id_event) REFERENCES event(id_event),
    FOREIGN KEY (user_id) REFERENCES "user"(id_user)
);

-- Tabela de logs de auditoria
CREATE TABLE audit_log (
    id_audit_log UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_tenant UUID,
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(10) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    user_id UUID,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_tenant) REFERENCES tenant(id_tenant),
    FOREIGN KEY (user_id) REFERENCES "user"(id_user)
);

-- Índices para melhorar performance
CREATE INDEX idx_user_id_tenant ON "user"(id_tenant);
CREATE INDEX idx_role_id_tenant ON role(id_tenant);
CREATE INDEX idx_event_id_tenant ON event(id_tenant);
CREATE INDEX idx_partner_id_tenant ON partner(id_tenant);
CREATE INDEX idx_employee_id_tenant ON employee(id_tenant);
CREATE INDEX idx_checkin_id_event ON checkin(id_event);
CREATE INDEX idx_checkout_id_event ON checkout(id_event);
CREATE INDEX idx_event_log_id_event ON event_log(id_event);
CREATE INDEX idx_audit_log_id_tenant ON audit_log(id_tenant);
CREATE INDEX idx_audit_log_table_name ON audit_log(table_name);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);

-- Índices para campos JSONB
CREATE INDEX idx_config_tenant_modules ON config_tenant USING GIN (modules);
CREATE INDEX idx_event_log_details ON event_log USING GIN (details);
CREATE INDEX idx_audit_log_old_values ON audit_log USING GIN (old_values);
CREATE INDEX idx_audit_log_new_values ON audit_log USING GIN (new_values);

-- Índices para QR Codes e métodos de check-in
CREATE INDEX idx_event_qr_code_token ON event_qr_code(qr_token);
CREATE INDEX idx_checkin_method ON checkin(check_method);
CREATE INDEX idx_checkout_method ON checkout(check_method);

-- Trigger para atualizar o campo updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tenant_updated_at BEFORE UPDATE ON tenant FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_user_updated_at BEFORE UPDATE ON "user" FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_role_updated_at BEFORE UPDATE ON role FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_event_updated_at BEFORE UPDATE ON event FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_partner_updated_at BEFORE UPDATE ON partner FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_employee_updated_at BEFORE UPDATE ON employee FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();