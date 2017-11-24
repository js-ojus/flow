-- This is only for testing purposes.  In production mode, the
-- consuming application is expected to create the database.
--
-- N.B.  The character set must be specified as 'utf8mb4' for proper
-- UTF-8 handling.  A simple 'utf8' does not suffice.

DROP DATABASE IF EXISTS flow;

--

CREATE DATABASE flow
CHARACTER SET = 'utf8mb4'
COLLATE = 'utf8mb4_unicode_ci';

SET collation_server = 'utf8mb4_unicode_ci';
SET collation_connection = 'utf8mb4_unicode_ci';
