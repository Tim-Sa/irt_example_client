-- +goose Up
CREATE TABLE subjects (
    id int NOT NULL,
    subject text,
    ability FLOAT,
    PRIMARY KEY(id)
);

-- +goose Down
DROP TABLE subjects;