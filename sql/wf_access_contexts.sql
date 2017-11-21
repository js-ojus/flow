DROP TABLE IF EXISTS wf_access_contexts;

--

CREATE TABLE wf_access_contexts (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    active TINYINT(1) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);
