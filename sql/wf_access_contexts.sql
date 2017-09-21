DROP TABLE IF EXISTS wf_access_contexts;

--

CREATE TABLE wf_access_contexts (
    id INT NOT NULL AUTO_INCREMENT,
    ns_id INT NOT NULL,
    group_id INT NOT NULL,
    role_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
    FOREIGN KEY (role_id) REFERENCES wf_roles_master(id)
);
