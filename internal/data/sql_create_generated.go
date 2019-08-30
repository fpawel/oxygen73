package data

const SQLCreate = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS product_voltage
(
    stored_at         REAL    NOT NULL,
    series_created_at REAL    NOT NULL,
    place             INTEGER NOT NULL CHECK (place >= 0 AND place <= 49),
    serial_number     INTEGER NOT NULL CHECK (serial_number > 0),
    tension           REAL    NOT NULL,
    PRIMARY KEY (stored_at, place, serial_number)
);

CREATE TABLE IF NOT EXISTS ambient
(
    stored_at         REAL NOT NULL PRIMARY KEY,
    series_created_at REAL NOT NULL,
    temperature       REAL NOT NULL,
    pressure          REAL NOT NULL,
    humidity          REAL NOT NULL
);

DROP VIEW IF EXISTS product_voltage_time;
CREATE VIEW IF NOT EXISTS product_voltage_time AS
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', series_created_at) as series_created_at_str,
       STRFTIME('%Y-%m-%d %H:%M:%f', stored_at ) as stored_at_str,
       CAST( (julianday('now', 'localtime') - stored_at) * 86400. / 60. AS INTEGER) AS minutes_elapsed,
       *
FROM product_voltage;

DROP VIEW IF EXISTS product_voltage_updated_at;
CREATE VIEW IF NOT EXISTS product_voltage_updated_at AS
SELECT *
FROM product_voltage_time
ORDER BY stored_at DESC
LIMIT 1;

DROP VIEW IF EXISTS product_voltage_series;
CREATE VIEW IF NOT EXISTS product_voltage_series AS
    WITH q1 AS (
        SELECT series_created_at AS started_at,
               max(stored_at) AS updated_at,
               (max(stored_at) - min(stored_at)) * 86400. AS total_seconds
        FROM product_voltage_time
        GROUP BY series_created_at
    )
    SELECT started_at,
           updated_at,
           strftime('%Y-%m-%d %H:%M:%f', started_at) AS started_at_str,
           strftime('%Y-%m-%d %H:%M:%f', updated_at) AS updated_at_str,
           CAST( total_seconds / 60. AS INTEGER) AS minutes,
           CAST( total_seconds % 60. AS INTEGER) AS seconds,
           cast(strftime('%Y', started_at) AS INTEGER) AS year,
           cast(strftime('%m', started_at) AS INTEGER) AS month,
           cast(strftime('%d', started_at) AS INTEGER) AS day
    FROM q1;



--SELECT CAST( (julianday('2019-08-29 23:45:55') -julianday('2019-08-29 23:38:03')) * 86400. AS INTEGER);`
