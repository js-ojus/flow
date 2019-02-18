DROP TABLE IF EXISTS wf_groups;

--

CREATE TABLE wf_groups (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    group_type ENUM('G', 'S'),
    PRIMARY KEY (id),
    UNIQUE (name)
);
