package data

const SQLCreate = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS product_voltage
(
    place   INTEGER   NOT NULL CHECK (place >= 0 AND place <= 50),
    number  INTEGER   NOT NULL CHECK (number > 0),
    time    TIMESTAMP NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW', 'localtime')),
    tension REAL      NOT NULL,
    PRIMARY KEY (place, number, time)
);

CREATE TABLE IF NOT EXISTS ambient
(
    time        TIMESTAMP NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW', 'localtime')) PRIMARY KEY,
    temperature REAL      NOT NULL,
    pressure    REAL      NOT NULL,
    humidity    REAL      NOT NULL
);`
