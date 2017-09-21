DROP TABLE IF EXISTS wf_docevent_application;

--

CREATE TABLE wf_docevent_application (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    doc_id INT NOT NULL,
    from_state_id INT NOT NULL,
    docevent_id INT NOT NULL,
    to_state_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    FOREIGN KEY (from_state_id) REFERENCES wf_docstates_master(id),
    FOREIGN KEY (docevent_id) REFERENCES wf_docevents(id),
    FOREIGN KEY (to_state_id) REFERENCES wf_docstates_master(id)
);
