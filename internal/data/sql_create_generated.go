package data

const SQLCreate = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS product_voltage
(
    stored_at     REAL    NOT NULL,
    place         INTEGER NOT NULL CHECK (place >= 0 AND place <= 49),
    serial_number INTEGER NOT NULL CHECK (serial_number > 0),
    tension       REAL    NOT NULL,
    PRIMARY KEY (stored_at, place, serial_number)
);

CREATE TABLE IF NOT EXISTS ambient
(
    stored_at   REAL NOT NULL PRIMARY KEY,
    temperature REAL NOT NULL,
    pressure    REAL NOT NULL,
    humidity    REAL NOT NULL
);

DROP VIEW IF EXISTS product_voltage_time;
CREATE VIEW IF NOT EXISTS product_voltage_time AS
    WITH v1 AS (
        SELECT *,
               cast(strftime('%Y', stored_at) AS INTEGER) AS year,
               cast(strftime('%m', stored_at) AS INTEGER) AS month,
               cast(strftime('%d', stored_at) AS INTEGER) AS day,
               cast(strftime('%H', stored_at) AS INTEGER) AS hour,
               cast(strftime('%M', stored_at) AS INTEGER) AS minute,
               cast(strftime('%f', stored_at) AS REAL)    AS second_real
        FROM product_voltage
    ),
         v2 AS (
             SELECT *, cast(second_real AS INTEGER) AS second
             FROM v1
         )
    SELECT *,
           cast((second_real - second) * 1000 AS INTEGER) AS millisecond
    FROM v2;

DROP VIEW product_voltage_updated_at;
CREATE VIEW IF NOT EXISTS product_voltage_updated_at AS
SELECT year, month, day, hour, minute, second, millisecond
FROM product_voltage_time
ORDER BY stored_at DESC
LIMIT 1;

`
