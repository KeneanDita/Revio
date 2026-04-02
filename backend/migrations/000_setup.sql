-- Run as postgres superuser
-- Grants the revio user full access to its database schema

\c revio

GRANT ALL PRIVILEGES ON DATABASE revio TO revio;
GRANT ALL ON SCHEMA public TO revio;
ALTER SCHEMA public OWNER TO revio;
