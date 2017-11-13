DROP TABLE IF EXISTS wf_ac_group_roles;

--

CREATE TABLE wf_ac_group_roles (
    id INT NOT NULL AUTO_INCREMENT,
    ac_id INT NOT NULL,
    group_id INT NOT NULL,
    role_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (ac_id) REFERENCES wf_access_contexts(id),
    FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
    FOREIGN KEY (role_id) REFERENCES wf_roles_master(id)
);
