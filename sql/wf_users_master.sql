-- This assumes the existence of a master table for users, by name
-- `users_master`.  It also assumes availability of the specified
-- columns in that master table.  This may need to be edited
-- appropriately, depending on your application and database design.

CREATE OR REPLACE VIEW wf_users_master AS
SELECT id, first_name, last_name, email, status
FROM users_master;
