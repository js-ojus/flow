DROP TABLE IF EXISTS wf_docstate_transitions;

--

CREATE TABLE wf_docstate_transitions (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    from_state_id INT NOT NULL,
    docaction_id INT NOT NULL,
    to_state_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes(id),
    FOREIGN KEY (from_state_id) REFERENCES wf_docstates(id),
    FOREIGN KEY (docaction_id) REFERENCES wf_docactions(id),
    FOREIGN KEY (to_state_id) REFERENCES wf_docstates(id),
    UNIQUE (doctype_id, from_state_id, docaction_id, to_state_id)
);
