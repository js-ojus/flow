CREATE TABLE wf_workflow_nodes (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    docstate_id INT NOT NULL,
    workflow_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    type ENUM('begin', 'end', 'linear', 'branch', 'joinany', 'joinall') NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    FOREIGN KEY (workflow_id) REFERENCES wf_workflows(id),
    FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id),
    UNIQUE (doctype_id, docstate_id)
);
