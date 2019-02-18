-- This assumes the existence of a master table for users, by name
-- `users`.  It also assumes availability of the specified columns in
-- that master table.  This may need to be edited appropriately,
-- depending on your application and database design.

CREATE OR REPLACE VIEW wf_users AS
SELECT id, first_name, last_name, email, active
FROM users;
