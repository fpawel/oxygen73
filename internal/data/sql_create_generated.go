package data

const SQLCreate = `
PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';

CREATE TABLE IF NOT EXISTS party
(
    party_id   INTEGER PRIMARY KEY NOT NULL,
    created_at TIMESTAMP           NOT NULL DEFAULT (DATETIME('now'))
);

CREATE VIEW IF NOT EXISTS last_party AS
SELECT *
FROM party
ORDER BY created_at DESC
LIMIT 1;

CREATE TABLE IF NOT EXISTS product
(
    product_id   INTEGER PRIMARY KEY NOT NULL,
    party_id     INTEGER             NOT NULL,
    serial       SMALLINT            NOT NULL CHECK (serial > 0 ),
    place        SMALLINT            NOT NULL CHECK (place >= 0 AND place < 50),
    product_type SMALLINT            NOT NULL CHECK (product_type > 0),
    UNIQUE (party_id, place),
    UNIQUE (party_id, serial),
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS measurement
(
    stored_at   REAL NOT NULL PRIMARY KEY,
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

CREATE VIEW IF NOT EXISTS last_bucket AS
SELECT *
FROM bucket
ORDER BY created_at DESC
LIMIT 1;

CREATE TABLE IF NOT EXISTS bucket
(
    created_at TIMESTAMP NOT NULL PRIMARY KEY DEFAULT (datetime('now')),
    updated_at TIMESTAMP NOT NULL             DEFAULT (datetime('now')),
    party_id   INTEGER   NOT NULL,
    FOREIGN KEY (party_id) REFERENCES party (party_id) ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS trigger_bucket_insert
    AFTER INSERT
    ON measurement
    WHEN NOT EXISTS(SELECT created_at
                    FROM bucket) OR
         (new.stored_at - julianday((SELECT updated_at
                                     FROM last_bucket))) * 86400. / 60. > 5
        OR (SELECT party_id
            FROM last_party) != (SELECT party_id
                                 FROM last_bucket)

BEGIN
    INSERT INTO bucket (created_at, updated_at, party_id)
    VALUES (datetime(new.stored_at), datetime(new.stored_at), (SELECT party_id FROM last_party));
END;

CREATE TRIGGER IF NOT EXISTS trigger_bucket_update
    AFTER INSERT
    ON measurement
    WHEN (new.stored_at - julianday((SELECT updated_at
                                     FROM bucket
                                     ORDER BY created_at DESC
                                     LIMIT 1))) * 86400. / 60. < 5
BEGIN
    UPDATE bucket
    SET updated_at = DATETIME(new.stored_at)
    WHERE created_at = (SELECT created_at
                        FROM bucket
                        ORDER BY created_at DESC
                        LIMIT 1);
END;

-- CAST((julianday('now', 'localtime') - stored_at) * 86400. / 60. AS INTEGER) AS minutes_elapsed,

--SELECT CAST( (julianday('2019-08-29 23:45:55') -julianday('2019-08-29 23:38:03')) * 86400. AS INTEGER);`
