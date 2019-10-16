PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party
(
    party_id   INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL UNIQUE DEFAULT (DATETIME('now'))
);

CREATE VIEW IF NOT EXISTS last_party AS
SELECT *
FROM party
ORDER BY created_at DESC
LIMIT 1;

CREATE TABLE IF NOT EXISTS product
(
    product_id INTEGER PRIMARY KEY NOT NULL,
    party_id   INTEGER             NOT NULL,
    serial     SMALLINT            NOT NULL CHECK (serial >= 0 ),
    place      SMALLINT            NOT NULL CHECK (place >= 0 AND place < 50),
    UNIQUE (party_id, place),
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS index_product_serial ON product (serial);

CREATE TABLE IF NOT EXISTS bucket
(
    bucket_id  INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL UNIQUE DEFAULT (datetime('now')),
    updated_at TIMESTAMP           NOT NULL        DEFAULT (datetime('now')),
    party_id   INTEGER             NOT NULL,
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE
);

CREATE VIEW IF NOT EXISTS last_bucket AS
SELECT *
FROM bucket
ORDER BY created_at DESC
LIMIT 1;

CREATE TABLE IF NOT EXISTS measurement
(
    tm          REAL    NOT NULL PRIMARY KEY,
    place0      REAL,
    place1      REAL,
    place2      REAL,
    place3      REAL,
    place4      REAL,
    place5      REAL,
    place6      REAL,
    place7      REAL,
    place8      REAL,
    place9      REAL,
    place10     REAL,
    place11     REAL,
    place12     REAL,
    place13     REAL,
    place14     REAL,
    place15     REAL,
    place16     REAL,
    place17     REAL,
    place18     REAL,
    place19     REAL,
    place20     REAL,
    place21     REAL,
    place22     REAL,
    place23     REAL,
    place24     REAL,
    place25     REAL,
    place26     REAL,
    place27     REAL,
    place28     REAL,
    place29     REAL,
    place30     REAL,
    place31     REAL,
    place32     REAL,
    place33     REAL,
    place34     REAL,
    place35     REAL,
    place36     REAL,
    place37     REAL,
    place38     REAL,
    place39     REAL,
    place40     REAL,
    place41     REAL,
    place42     REAL,
    place43     REAL,
    place44     REAL,
    place45     REAL,
    place46     REAL,
    place47     REAL,
    place48     REAL,
    place49     REAL,
    temperature REAL,
    pressure    REAL,
    humidity    REAL
);

CREATE TRIGGER IF NOT EXISTS trigger_bucket_insert
    AFTER INSERT
    ON measurement
    WHEN NOT EXISTS(SELECT created_at
                    FROM bucket) OR
         (new.tm - julianday((SELECT updated_at
                              FROM last_bucket))) * 86400. / 60. > 5
        OR (SELECT party_id
            FROM last_party) != (SELECT party_id
                                 FROM last_bucket)
BEGIN
    INSERT INTO bucket (created_at, updated_at, party_id)
    VALUES (datetime(new.tm), datetime(new.tm), (SELECT party_id FROM last_party));
END;

CREATE TRIGGER IF NOT EXISTS trigger_bucket_update
    AFTER INSERT
    ON measurement
    WHEN (new.tm - julianday((SELECT updated_at FROM last_bucket))) * 86400. / 60. < 5
BEGIN
    UPDATE bucket
    SET updated_at = DATETIME(new.tm)
    WHERE bucket_id = (SELECT bucket_id FROM last_bucket);
END;

CREATE VIEW IF NOT EXISTS measurement1 AS
SELECT STRFTIME('%Y-%m-%d %H:%M:%f', tm) AS stored_at,
       cast(strftime('%Y', tm) AS INTEGER) AS year,
       cast(strftime('%m', tm) AS INTEGER) AS month,
       *
FROM measurement;

CREATE VIEW IF NOT EXISTS bucket1 AS
SELECT bucket.*,
       party.created_at AS party_created_at,
       cast(strftime('%Y', bucket.created_at) AS INTEGER) AS year,
       cast(strftime('%m', bucket.created_at) AS INTEGER) AS month,
       bucket_id = (SELECT bucket_id FROM last_bucket) AS is_last
FROM bucket
INNER JOIN party USING (party_id)
ORDER BY bucket.created_at;

-- CREATE VIEW measurement_ids AS
-- SELECT (SELECT count(*) + 1 FROM measurement WHERE tm < O.tm) AS id, O.*
-- FROM measurement O
-- ORDER BY 2;
--
--
-- DROP VIEW IF EXISTS start_finish;
-- CREATE VIEW start_finish AS
--     SELECT O1.tm AS tm1,
--            O2.tm AS tm2
--     FROM measurement_ids O1
--              INNER JOIN measurement_ids O2
--                         ON O1.id = O2.id - 1
--     WHERE O1.party_id != O2.party_id
--        OR abs(tm1 - tm2) * 86400. / 60. > 5
--     UNION
--     SELECT NULL, tm
--     FROM measurement
--     WHERE tm = (SELECT min(tm) FROM measurement)
--     UNION
--     SELECT tm, NULL
--     FROM measurement
--     WHERE tm = (SELECT max(tm) FROM measurement);
--
-- DROP VIEW IF EXISTS measurement_start_finish;
--
-- CREATE VIEW measurement_start_finish AS
--     WITH Q1 AS (
--         SELECT O3.*,
--                (SELECT max(tm2)
--                 FROM start_finish
--                 WHERE tm2 <= O3.tm) AS start,
--                (SELECT min(tm1)
--                 FROM start_finish
--                 WHERE tm1 >= O3.tm) AS finish
--         FROM measurement_ids O3)
--     SELECT *,
--            STRFTIME('%Y-%m-%d %H:%M:%f', tm)     AS stored_at,
--            STRFTIME('%Y-%m-%d %H:%M:%f', start)  AS start_at,
--            STRFTIME('%Y-%m-%d %H:%M:%f', finish) AS finish_at,
--            cast(strftime('%Y', tm) AS INTEGER)   AS year,
--            cast(strftime('%m', tm) AS INTEGER)   AS month
--     FROM Q1
--     ORDER BY stored_at;

-- DROP VIEW IF EXISTS last_series;
-- CREATE VIEW last_series AS
-- SELECT *
-- FROM measurement_start_finish
-- ORDER BY start DESC
-- LIMIT 1;

-- CAST((julianday('now', 'localtime') - stored_at) * 86400. / 60. AS INTEGER) AS minutes_elapsed,

--SELECT CAST( (julianday('2019-08-29 23:45:55') -julianday('2019-08-29 23:38:03')) * 86400. AS INTEGER);