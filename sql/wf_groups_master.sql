CREATE TABLE wf_groups_master (
id INT PRIMARY KEY,
name TEXT NOT NULL,
group_type CHAR(1),
FOREIGN KEY (name) REFERENCES users_master(email),
UNIQUE (name)
);
