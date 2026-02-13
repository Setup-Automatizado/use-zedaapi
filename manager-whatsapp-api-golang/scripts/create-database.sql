-- ==================================================
-- Manager Database Setup Script
-- ==================================================
-- Execute este script no RDS existente para criar
-- o banco de dados do Manager
-- ==================================================

-- Criar banco de dados do Manager (se nao existir)
-- NOTA: Execute como superuser ou role com CREATEDB
CREATE DATABASE manager_db
    WITH
    OWNER = whatsapp_api_user
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TEMPLATE = template0;

-- Conceder privilegios
GRANT ALL PRIVILEGES ON DATABASE manager_db TO whatsapp_api_user;

-- Conectar ao banco manager_db e criar schema
\c manager_db

-- Criar extensoes necessarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Comentario
COMMENT ON DATABASE manager_db IS 'Manager WhatsApp API - Frontend authentication and user management';
