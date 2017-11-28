-- CREATE TABLE wf_documents_<DOCTYPE_ID> (
--     id INT NOT NULL AUTO_INCREMENT,
--     path VARCHAR(1000) NOT NULL,
--     ac_id INT NOT NULL,
--     docstate_id INT NOT NULL,
--     group_id INT NOT NULL,
--     ctime TIMESTAMP NOT NULL,
--     title VARCHAR(250) NULL,
--     data TEXT NOT NULL,
--     PRIMARY KEY (id),
--     FOREIGN KEY (ac_id) REFERENCES wf_access_contexts(id),
--     FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id),
--     FOREIGN KEY (group_id) REFERENCES wf_groups_master(id)
-- );

--

DROP TABLE IF EXISTS wf_document_children;

CREATE TABLE wf_document_children (
    id INT NOT NULL AUTO_INCREMENT,
    parent_doctype_id INT NOT NULL,
    parent_id INT NOT NULL,
    child_doctype_id INT NOT NULL,
    child_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (parent_doctype_id) REFERENCES wf_doctypes_master(id),
    FOREIGN KEY (child_doctype_id) REFERENCES wf_doctypes_master(id),
    UNIQUE (parent_doctype_id, parent_id, child_doctype_id, child_id)
);

--

DROP TABLE IF EXISTS wf_document_blobs;

CREATE TABLE wf_document_blobs (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    doc_id INT NOT NULL,
    sha1sum CHAR(40) NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    UNIQUE (doctype_id, doc_id, sha1sum)
);

--

DROP TABLE IF EXISTS wf_document_tags;

CREATE TABLE wf_document_tags (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    doc_id INT NOT NULL,
    tag VARCHAR(50) NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    UNIQUE (doctype_id, doc_id, tag)
);
