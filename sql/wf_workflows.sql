DROP TABLE IF EXISTS wf_workflows;

--

CREATE TABLE wf_workflows (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    docstate_id INT NOT NULL,
    active TINYINT(1) NOT NULL,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes(id),
    FOREIGN KEY (docstate_id) REFERENCES wf_docstates(id),
    UNIQUE (name)
);
