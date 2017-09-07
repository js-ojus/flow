CREATE OR REPLACE VIEW wf_users_master AS
SELECT id, first_name, last_name, email, status
FROM users_master;
