DROP TABLE IF EXISTS wf_doctypes;

--

CREATE TABLE wf_doctypes (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (name)
);
