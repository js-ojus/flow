DROP TABLE IF EXISTS wf_ac_user_levels;

--

CREATE TABLE wf_ac_user_levels(
    id INT NOT NULL AUTO_INCREMENT,
    ac_id INT NOT NULL,
    user_id INT NOT NULL,
    ulevel INT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (ac_id) REFERENCES wf_access_contexts(id),
    UNIQUE (ac_id, user_id)
);
