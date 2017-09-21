CREATE TABLE IF NOT EXISTS wf_docstates_master (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    name TEXT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    UNIQUE (doctype_id, name)
);
