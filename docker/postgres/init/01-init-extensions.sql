-- ==============================================================================
-- Strata Database Initialization: Extensions Setup
-- ==============================================================================

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable cryptographic functions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enable vector search capabilities (pgvector)
CREATE EXTENSION IF NOT EXISTS "vector";

-- Enable geographic and spatial features (PostGIS)
CREATE EXTENSION IF NOT EXISTS "postgis";
