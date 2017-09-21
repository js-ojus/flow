DROP TABLE IF EXISTS wf_docactions_master;

--

CREATE TABLE wf_docactions_master (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);
