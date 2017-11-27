DROP TABLE IF EXISTS wf_docevents;

--

CREATE TABLE wf_docevents (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    doc_id INT NOT NULL,
    docstate_id INT NOT NULL,
    docaction_id INT NOT NULL,
    group_id INT NOT NULL,
    data TEXT,
    ctime TIMESTAMP NOT NULL,
    status ENUM('A', 'P') NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id),
    FOREIGN KEY (docaction_id) REFERENCES wf_docactions_master(id),
    FOREIGN KEY (group_id) REFERENCES wf_groups_master(id)
);
