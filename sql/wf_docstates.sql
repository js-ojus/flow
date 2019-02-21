DROP TABLE IF EXISTS wf_docstates;

--

CREATE TABLE wf_docstates (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);

--

-- This reserved state has ID `1`.  This is used as the initial state of
-- documents that are created without an explicit state.
INSERT INTO wf_docstates(name)
VALUES('__INITIAL__');
