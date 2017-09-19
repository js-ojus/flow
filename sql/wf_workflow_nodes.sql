CREATE TABLE wf_workflow_nodes (
    id INT NOT NULL AUTO_INCREMENT,
    workflow_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    type ENUM('begin', 'end', 'linear', 'branch', 'joinany', 'joinall') NOT NULL,
    docstate_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (workflow_id) REFERENCES wf_workflows(id),
    FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id),
    UNIQUE (workflow_id, name)
);
