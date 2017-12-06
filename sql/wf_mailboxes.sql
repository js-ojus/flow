DROP TABLE IF EXISTS wf_mailboxes;

--

CREATE TABLE wf_mailboxes (
    id INT NOT NULL AUTO_INCREMENT,
    group_id INT NOT NULL,
    message_id INT NOT NULL,
    unread TINYINT(1) NOT NULL,
    ctime TIMESTAMP NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (group_id) REFERENCES wf_groups_master(id),
    FOREIGN KEY (message_id) REFERENCES wf_messages(id),
    UNIQUE (group_id, message_id)
);
