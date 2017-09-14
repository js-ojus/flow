CREATE TABLE wf_docevents (
id INT NOT NULL AUTO_INCREMENT,
doc_id INT NOT NULL,
docstate_id INT NOT NULL,
user_id INT NOT NULL,
docaction_id INT NOT NULL,
data TEXT,
ctime TIMESTAMP NOT NULL,
applied BOOLEAN NOT NULL,
PRIMARY KEY (id),
FOREIGN KEY (doc_id) REFERENCES wf_documents(id),
FOREIGN KEY (docstate_id) REFERENCES wf_docstates_master(id),
FOREIGN KEY (user_id) REFERENCES users_masters(id),
FOREIGN KEY (docaction_id) REFERENCES wf_docactions_master(id)
);
