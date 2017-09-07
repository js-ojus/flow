CREATE TABLE wf_docstates_master (
id INT PRIMARY KEY,
doctype_id INT NOT NULL,
name TEXT NOT NULL,
FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
UNIQUE (doctype_id, name)
);
