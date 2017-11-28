DROP TABLE IF EXISTS wf_docstates_master;

--

CREATE TABLE wf_docstates_master (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);

--

-- This reserved state has ID `1`.  This is used as the only legal
-- state for children documents.
INSERT INTO wf_docstates_master(name)
VALUES('__RESERVED_CHILD_STATE__');
