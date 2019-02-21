DROP TABLE IF EXISTS wf_workflow_nodes;

--

CREATE TABLE wf_workflow_nodes (
    id INT NOT NULL AUTO_INCREMENT,
    workflow_id INT NOT NULL,
    doctype_id INT NOT NULL,
    docstate_id INT NOT NULL, -- document must be in this state for this node to get activated
    inherit_ac BOOLEAN NOT NULL,
    ac_id INT,
    node_type VARCHAR(20) NOT NULL, -- values come from `nodetype.go`
    PRIMARY KEY (id),
    FOREIGN KEY (workflow_id) REFERENCES wf_workflows(id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes(id),
    FOREIGN KEY (docstate_id) REFERENCES wf_docstates(id),
    FOREIGN KEY (ac_id) REFERENCES wf_access_contexts(id),
    UNIQUE (doctype_id, docstate_id),
    UNIQUE (workflow_id, name)
);
