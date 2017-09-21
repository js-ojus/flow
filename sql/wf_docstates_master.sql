DROP TABLE IF EXISTS wf_docstates_master;

--

CREATE TABLE wf_docstates_master (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    UNIQUE (doctype_id, name)
);
