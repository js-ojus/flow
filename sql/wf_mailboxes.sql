CREATE TABLE wf_mailboxes (
    id INT NOT NULL AUTO_INCREMENT,
    group_id INT NOT NULL,
    message_id INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
    FOREIGN KEY (message_id) REFERENCES wf_messages(id),
);
