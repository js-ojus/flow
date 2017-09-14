CREATE TABLE wf_groups_master (
id INT NOT NULL AUTO_INCREMENT,
name TEXT NOT NULL,
group_type CHAR(1),
PRIMARY KEY (id),
FOREIGN KEY (name) REFERENCES users_master(email),
UNIQUE (name)
);
