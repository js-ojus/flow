CREATE TABLE IF NOT EXISTS wf_groups_master (
    id INT NOT NULL AUTO_INCREMENT,
    name TEXT NOT NULL,
    group_type CHAR(1),
    PRIMARY KEY (id),
    UNIQUE (name)
);
