DROP TABLE IF EXISTS wf_docstates_master;

--

CREATE TABLE wf_docstates_master (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);
