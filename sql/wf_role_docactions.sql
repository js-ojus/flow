CREATE TABLE wf_role_docactions (
id INT PRIMARY KEY,
role_id INT NOT NULL,
doctype_id INT NOT NULL,
docaction_id INT NOT NULL,
FOREGIN KEY (role_id) REFERENCES wf_roles_master(id),
FOREIGN KEY (doctype_id) REFERENCES wf_doctypes_master(id),
FOREIGN KEY (docaction_id) REFERENCES wf_docactions_master(id)
);