#!/bin/sh
set -eu
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
 -v owner_pass="$OWNER_PASSWORD" -v agent_pass="$AGENT_PASSWORD" -v worker_pass="$WORKER_PASSWORD" -v operator_pass="$OPERATOR_PASSWORD" <<'SQL'
CREATE ROLE trading_owner LOGIN PASSWORD :'owner_pass' NOSUPERUSER NOBYPASSRLS;
CREATE ROLE trading_agent LOGIN PASSWORD :'agent_pass' NOSUPERUSER NOBYPASSRLS;
CREATE ROLE trading_worker LOGIN PASSWORD :'worker_pass' NOSUPERUSER NOBYPASSRLS;
CREATE ROLE trading_operator LOGIN PASSWORD :'operator_pass' NOSUPERUSER NOBYPASSRLS;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE,CREATE ON SCHEMA public TO trading_owner;
GRANT USAGE ON SCHEMA public TO trading_agent,trading_worker,trading_operator;
SQL
