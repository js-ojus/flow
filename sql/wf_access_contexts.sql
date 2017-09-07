CREATE TABLE wf_access_contexts (
id INT PRIMARY KEY,
ns_id INT NOT NULL,
group_id INT NOT NULL,
role_id INT NOT NULL,
FOREIGN KEY (ns_id) REFERENCES projects(id),
FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
FOREGIN KEY (role_id) REFERENCES wf_roles_master(id)
);
