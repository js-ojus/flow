DROP TABLE IF EXISTS wf_messages;

--

CREATE TABLE wf_messages (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    doc_id INT NOT NULL,
    docevent_id INT NOT NULL,
    title VARCHAR(250) NOT NULL,
    data TEXT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    FOREIGN KEY (docevent_id) REFERENCES wf_docevents(id),
    UNIQUE (doctype_id, doc_id, docevent_id)
);
