CREATE TABLE wf_workflow_nodes (
    id INT NOT NULL AUTO_INCREMENT,
    workflow_id INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    type ENUM('begin', 'end', 'linear', 'branch', 'joinany', 'joinall'),
    docstate_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (workflow_id) REFERENCES wf_workflows(id),
    FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id),
    UNIQUE (workflow_id, name)
);

--

CREATE TABLE wf_node_next_states (
    id INT NOT NULL AUTO_INCREMENT,
    node_id INT NOT NULL,
    docstate_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (node_id) REFERENCES wf_workflow_nodes(id),
    FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id)
);
