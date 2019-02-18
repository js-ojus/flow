DROP TABLE IF EXISTS wf_docactions;

--

CREATE TABLE wf_docactions (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    reconfirm TINYINT(1) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);
