CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    release_date DATE DEFAULT NULL,
    text TEXT,
    link VARCHAR(255),
    title varchar(128) NOT NULL ,
    group_name VARCHAR(255) NOT NULL
);
