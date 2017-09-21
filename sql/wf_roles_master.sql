DROP TABLE IF EXISTS wf_roles_master;

--

CREATE TABLE wf_roles_master (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);
