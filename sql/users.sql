DROP TABLE IF EXISTS users;

--

CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT,
    first_name VARCHAR(30) NOT NULL,
    last_name VARCHAR(30) NOT NULL,
    email VARCHAR(100) NOT NULL,
    active TINYINT(1) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (email)
);
