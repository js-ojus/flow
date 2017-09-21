-- CREATE TABLE IF NOT EXISTS wf_documents_<DOCTYPE_ID> (
--     id INT NOT NULL AUTO_INCREMENT,
--     user_id INT NOT NULL,
--     docstate_id INT NOT NULL,
--     ctime TIMESTAMP NOT NULL,
--     title VARCHAR(250) NOT NULL,
--     data BLOB NOT NULL,
--     PRIMARY KEY (id),
--     FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id)
-- );

--

CREATE TABLE IF NOT EXISTS wf_document_children (
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

CREATE TABLE IF NOT EXISTS wf_document_blobs (
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

CREATE TABLE IF NOT EXISTS wf_document_tags (
    id INT NOT NULL AUTO_INCREMENT,
    doctype_id INT NOT NULL,
    doc_id INT NOT NULL,
    tag VARCHAR(50) NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
    UNIQUE (doctype_id, doc_id, tag)
);
