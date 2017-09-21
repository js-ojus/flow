CREATE TABLE IF NOT EXISTS wf_access_contexts (
    id INT NOT NULL AUTO_INCREMENT,
    ns_id INT NOT NULL,
    group_id INT NOT NULL,
    role_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
    FOREGIN KEY (role_id) REFERENCES wf_roles_master(id)
);
