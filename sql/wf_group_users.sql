DROP TABLE IF EXISTS wf_group_users;

--

CREATE TABLE wf_group_users (
    id INT NOT NULL AUTO_INCREMENT,
    group_id INT NOT NULL,
    user_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
    UNIQUE (group_id, user_id)
);
