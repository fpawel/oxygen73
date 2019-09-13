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

CREATE TABLE IF NOT EXISTS product_voltage
(
    product_id INTEGER NOT NULL,
    stored_at  REAL    NOT NULL,
    voltage    REAL    NOT NULL,
    PRIMARY KEY (stored_at, product_id),
    FOREIGN KEY (product_id) REFERENCES product (product_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS bucket
(
    created_at TIMESTAMP NOT NULL PRIMARY KEY DEFAULT (datetime('now')),
    updated_at TIMESTAMP NOT NULL             DEFAULT (datetime('now'))
);

CREATE TRIGGER IF NOT EXISTS trigger_bucket_insert
    AFTER INSERT
    ON product_voltage
    WHEN NOT EXISTS(SELECT created_at
                    FROM bucket) OR
         (julianday('now') - julianday((SELECT updated_at
                                        FROM bucket
                                        ORDER BY created_at DESC
                                        LIMIT 1))) * 86400. / 60. > 5

BEGIN
    INSERT INTO bucket (created_at, updated_at) VALUES (datetime('now'), datetime('now'));
END;

CREATE TRIGGER IF NOT EXISTS trigger_bucket_update
    AFTER INSERT
    ON product_voltage
    WHEN (julianday('now') - julianday((SELECT updated_at
                                        FROM bucket
                                        ORDER BY created_at DESC
                                        LIMIT 1))) * 86400. / 60. < 5
BEGIN
    UPDATE bucket
    SET updated_at = DATETIME('now')
    WHERE created_at = (SELECT created_at
                        FROM bucket
                        ORDER BY created_at DESC
                        LIMIT 1);
END;

CREATE TABLE IF NOT EXISTS ambient
(
    stored_at   REAL NOT NULL PRIMARY KEY,
    temperature REAL NOT NULL,
    pressure    REAL NOT NULL,
    humidity    REAL NOT NULL
);

-- CAST((julianday('now', 'localtime') - stored_at) * 86400. / 60. AS INTEGER) AS minutes_elapsed,

--SELECT CAST( (julianday('2019-08-29 23:45:55') -julianday('2019-08-29 23:38:03')) * 86400. AS INTEGER);