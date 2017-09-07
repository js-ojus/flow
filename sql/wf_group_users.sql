CREATE TABLE wf_group_users (
id INT PRIMARY KEY,
group_id INT NOT NULL,
user_id INT NOT NULL,
FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
FOREIGN KEY (user_id) REFERENCES users_master(id),
UNIQUE (group_id, user_id)
);
