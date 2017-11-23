DROP TABLE IF EXISTS wf_ac_group_hierarchy;

--

CREATE TABLE wf_ac_group_hierarchy (
    id INT NOT NULL AUTO_INCREMENT,
    ac_id INT NOT NULL,
    group_id INT NOT NULL,
    reports_to INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (ac_id) REFERENCES wf_access_contexts(id),
    FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
    FOREIGN KEY (reports_to) REFERENCES wf_groups_master(id),
    UNIQUE (ac_id, group_id)
);
