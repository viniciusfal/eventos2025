# 📊 Documentação do Banco de Dados - Sistema de Check-in em Eventos (Melhorado)

## 📑 Entidade: `event_log` (logs detalhados de eventos)
Histórico de logs de eventos.
- `id_event_log` (uuid, PK)
- `id_tenant` (uuid, FK → tenant)
- `id_event` (uuid, FK → event)
- `log_type` (string) → tipo de log (checkin, checkout, system, user_action)
- `entity_type` (string) → tipo de entidade relacionada (employee, partner, user)
- `entity_id` (uuid) → ID da entidade relacionada
- `action` (string) → ação realizada
- `details` (json) → detalhes adicionais da ação
- `user_id` (uuid, FK → user) → usuário responsável (quando aplicável)
- `ip_address` (string) → IP do usuário (quando aplicável)
- `user_agent` (string) → User agent (quando aplicável)
- `location` (GEOGRAPHY) → localização geográfica (quando aplicável)
- `created_at` (timestamp)
Representa a empresa/organização (multi-tenant).
- `id_tenant` (uuid, PK)  
- `id_config_tenant` (uuid, FK → config_tenant)  
- `name` (string)  
- `identity` (string) (aqui pode ser um CNPJ, CPF e etc...)  
- `type_identity` (string) (aqui mostrara o tipo de identy, se é um CNPJ, CPF ...)
- `email` (string)  
- `address` (string)  
- `active` (bool)  
- `created_at` (timestamp)  
- `updated_at` (timestamp)  
- `created_by` (uuid, FK → user)
- `updated_by` (uuid, FK → user)

---

## ⚙️ Entidade: `config_tenant`
Configurações específicas do tenant.
- `id_config_tenant` (uuid, PK)  
- `modules` (json) → configurações de módulos habilitados  
- `updated_at` (timestamp)  
- `updated_by` (uuid, FK → user)

---

## 👤 Entidade: `user`
Usuários do sistema.
- `id_user` (uuid, PK)  
- `id_tenant` (uuid, FK → tenant)  
- `full_name` (string)  
- `email` (string)  
- `phone` (string)  
- `username` (string)  
- `password_hash` (string)  
- `active` (bool)  
- `created_at` (timestamp)  
- `updated_at` (timestamp)  
- `created_by` (uuid, FK → user)
- `updated_by` (uuid, FK → user)

### 🔐 Entidade: `user_role` (tabela de relacionamento)
- `id_user` (uuid, FK → user)
- `id_role` (uuid, FK → role)
- `assigned_at` (timestamp)
- `assigned_by` (uuid, FK → user)
- PK composta: (id_user, id_role)

---

## 🛡️ Entidade: `role`
Perfis de acesso do usuário.
- `id_role` (uuid, PK)  
- `id_tenant` (uuid, FK → tenant)  
- `name` (string)  
- `description` (string)  
- `active` (bool)
- `created_at` (timestamp)
- `updated_at` (timestamp)

### 🔑 Entidade: `permission`
- `id_permission` (uuid, PK)  
- `name` (string)  
- `description` (string)  
- `module` (string) → módulo ao qual a permissão pertence

### 🔗 Entidade: `role_permission` (tabela de relacionamento)
- `id_role` (uuid, FK → role)
- `id_permission` (uuid, FK → permission)
- PK composta: (id_role, id_permission)

---

## 🧩 Entidade: `module`
Módulos do sistema.
- `id_module` (uuid, PK)  
- `name_module` (string)  
- `description` (string)
- `active` (bool)

---

## 🎟️ Entidade: `event`
Eventos cadastrados.
- `id_event` (uuid, PK)  
- `id_tenant` (uuid, FK → tenant)  
- `event_name` (string)  
- `location` (string)  
- `fence_event` (GEOGRAPHY)  (Aqui teremos as coordenados de onde o evento esta acontecendo)
- `initial_date` (timestamp)  
- `final_date` (timestamp)  
- `active` (bool)  
- `created_at` (timestamp)  
- `updated_at` (timestamp)
- `created_by` (uuid, FK → user)
- `updated_by` (uuid, FK → user)

---

## 🔗 Entidade: `event_partner` (tabela de relacionamento)
- `id_event` (uuid, FK → event)
- `id_partner` (uuid, FK → partner)
- `assigned_at` (timestamp)
- `assigned_by` (uuid, FK → user)
- PK composta: (id_event, id_partner)

---

## 📦 Entidade: `event_qr_code`
QR Codes gerados para eventos.
- `id_qr_code` (uuid, PK)
- `id_event` (uuid, FK → event)
- `qr_type` (string) → tipo de QR Code ("checkin" ou "checkout")
- `qr_token` (string) → token único para validação
- `valid_from` (timestamp) → data de início da validade
- `valid_until` (timestamp) → data de fim da validade
- `created_at` (timestamp)
- `created_by` (uuid, FK → user)
Histórico de logs de eventos.
- `id_event_log` (uuid, PK)
- `id_tenant` (uuid, FK → tenant)
- `id_event` (uuid, FK → event)
- `log_type` (string) → tipo de log (checkin, checkout, system, user_action)
- `entity_type` (string) → tipo de entidade relacionada (employee, partner, user)
- `entity_id` (uuid) → ID da entidade relacionada
- `action` (string) → ação realizada
- `details` (json) → detalhes adicionais da ação
- `user_id` (uuid, FK → user) → usuário responsável (quando aplicável)
- `ip_address` (string) → IP do usuário (quando aplicável)
- `user_agent` (string) → User agent (quando aplicável)
- `location` (GEOGRAPHY) → localização geográfica (quando aplicável)
- `created_at` (timestamp)

---

## 🏢 Entidade: `partner`
Parceiros vinculados ao evento.
- `id_partner` (uuid, PK)  
- `id_tenant` (uuid, FK → tenant)  
- `name_partner` (string)  
- `email` (string)  
- `email_2` (string)  
- `phone` (string)  
- `phone_2` (string)  
- `identity` (string)  
- `type_identity` (string)  
- `location` (string)  
- `pass_hash` (string)  
- `last_login` (timestamp)  
- `failed_login_attempts` (integer)  
- `locked_until` (timestamp)  
- `active` (bool)  
- `created_at` (timestamp)  
- `updated_at` (timestamp)
- `created_by` (uuid, FK → user)
- `updated_by` (uuid, FK → user)

---

## 🔗 Entidade: `partner_employee` (tabela de relacionamento)
- `id_partner` (uuid, FK → partner)
- `id_employee` (uuid, FK → employee)
- `assigned_at` (timestamp)
- `assigned_by` (uuid, FK → user)
- PK composta: (id_partner, id_employee)

---

## 👨‍💼 Entidade: `employee`
Funcionários ligados ao tenant ou parceiros.
- `id_employee` (uuid, PK)  
- `id_tenant` (uuid, FK → tenant)  
- `full_name` (string)  
- `identity` (string)  
- `type_identity` (string)  
- `date_of_birth` (date)  
- `photo_url` (string)  
- `face_embedding` (VECTOR[512]) → usado para reconhecimento facial  
- `phone` (string)  
- `email` (string)  
- `active` (bool)  
- `created_at` (timestamp)  
- `updated_at` (timestamp)
- `created_by` (uuid, FK → user)
- `updated_by` (uuid, FK → user)

---

## ✅ Entidade: `checkin`
Registros de entrada em eventos.
- `id_checkin` (uuid, PK)  
- `id_tenant` (uuid, FK → tenant)  
- `id_event` (uuid, FK → event)  
- `id_employee` (uuid, FK → employee)  
- `id_partner` (uuid, FK → partner)  
- `user_resp_checkin` (uuid, FK → user)  
- `checkin_date_time` (timestamp)  
- `location` (GEOGRAPHY)  
- `device_id` (string) → ID do dispositivo usado para o checkin
- `notes` (text) → observações adicionais
- `check_method` (string) → método utilizado para checkin ("facial_recognition", "qr_code", "manual")
- `created_at` (timestamp)

---

## 🚪 Entidade: `checkout`
Registros de saída de eventos.
- `id_checkout` (uuid, PK)  
- `id_tenant` (uuid, FK → tenant)  
- `id_event` (uuid, FK → event)  
- `id_employee` (uuid, FK → employee)  
- `id_partner` (uuid, FK → partner)  
- `user_resp_checkout` (uuid, FK → user)  
- `checkout_date_time` (timestamp)  
- `location_checkout` (GEOGRAPHY)  
- `device_id` (string) → ID do dispositivo usado para o checkout
- `notes` (text) → observações adicionais
- `check_method` (string) → método utilizado para checkout ("facial_recognition", "qr_code", "manual")
- `created_at` (timestamp)

---

## 📊 Entidade: `audit_log` (logs de auditoria gerais do sistema)
- `id_audit_log` (uuid, PK)
- `id_tenant` (uuid, FK → tenant)
- `table_name` (string) → nome da tabela afetada
- `record_id` (uuid) → ID do registro afetado
- `action` (string) → ação realizada (INSERT, UPDATE, DELETE)
- `old_values` (json) → valores anteriores (para UPDATE/DELETE)
- `new_values` (json) → valores novos (para INSERT/UPDATE)
- `user_id` (uuid, FK → user) → usuário responsável
- `ip_address` (string) → IP do usuário
- `user_agent` (string) → User agent
- `created_at` (timestamp)

---

# 🔗 Principais Relacionamentos
- **Tenant** é a base multi-empresa (multi-tenant).  
- **User** pertence a um **Tenant**, com permissões e papéis (Role/Permission).  
- **Event** é vinculado a um **Tenant**, podendo ter **Partners** e **Employees** através de tabelas de relacionamento.  
- **Partner** pode ter vários **Employees** associados através da tabela `partner_employee`.  
- **Checkin/Checkout** registram a movimentação de **Employees** e **Partners** em um **Event**.  
- **Event_log** centraliza histórico detalhado de movimentações e ações no evento.  
- **Audit_log** registra todas as mudanças importantes no sistema para auditoria.  
- **Employee** suporta reconhecimento facial via `face_embedding`.
- **Event_qr_code** é vinculado a um **Event** para checkins/checkouts via QR Code.