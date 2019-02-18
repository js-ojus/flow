DROP TABLE IF EXISTS wf_role_docactions;

--

CREATE TABLE wf_role_docactions (
    id INT NOT NULL AUTO_INCREMENT,
    role_id INT NOT NULL,
    doctype_id INT NOT NULL,
    docaction_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (role_id) REFERENCES wf_roles(id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes(id),
    FOREIGN KEY (docaction_id) REFERENCES wf_docactions(id),
    UNIQUE (role_id, doctype_id, docaction_id)
);
