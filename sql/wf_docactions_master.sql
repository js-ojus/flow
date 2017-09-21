CREATE TABLE IF NOT EXISTS wf_docactions_master (
    id INT NOT NULL AUTO_INCREMENT,
    name TEXT NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);
