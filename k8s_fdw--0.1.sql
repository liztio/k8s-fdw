/* k8s-fdw/k8s-fdw--0.1.sql */

-- complain if script is sourced in psql, rather than via CREATE EXTENSION
\echo Use "CREATE EXTENSION go_fdw" to load this extension. \quit

CREATE FUNCTION go_fdw_handler()
RETURNS fdw_handler
AS 'MODULE_PATHNAME'
LANGUAGE C STRICT;

CREATE FUNCTION go_fdw_validator(text[], oid)
RETURNS void
AS 'MODULE_PATHNAME'
LANGUAGE C STRICT;

CREATE FOREIGN DATA WRAPPER k8s_fdw
  HANDLER go_fdw_handler
  VALIDATOR go_fdw_validator;
