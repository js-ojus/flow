DROP TABLE IF EXISTS wf_docstate_transitions;

--

CREATE TABLE wf_docstate_transitions (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    from_state_id INT NOT NULL,
    docaction_id INT NOT NULL,
    to_state_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    FOREIGN KEY (from_state_id) REFERENCES wf_docstates_master(id),
    FOREIGN KEY (docaction_id) REFERENCES wf_docactions_master(id),
    FOREIGN KEY (to_state_id) REFERENCES wf_docstates_master(id),
    UNIQUE (doctype_id, from_state_id, docaction_id, to_state_id)
);
