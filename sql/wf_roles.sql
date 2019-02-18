DROP TABLE IF EXISTS wf_roles;

--

CREATE TABLE wf_roles (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);

--

-- This reserved role is for users who should administer `flow`
-- itself. That includes (but is not limited to) definition and
-- management of document types, their workflows, roles and groups.
INSERT INTO wf_roles(name)
VALUES('SUPER_ADMIN');

-- This reserved role is for users who assume apex positions in
-- day-to-day operations.  This role can be used to administer the
-- workflow operations within access contexts, when needed.
INSERT INTO wf_roles(name)
VALUES('ADMIN');
