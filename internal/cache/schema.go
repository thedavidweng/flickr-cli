package cache

// Schema defines the SQLite cache tables.
const Schema = `
CREATE TABLE IF NOT EXISTS profiles (
    profile TEXT PRIMARY KEY,
    user_id TEXT,
    username TEXT,
    updated_at TEXT
);

CREATE TABLE IF NOT EXISTS albums (
    profile TEXT,
    id TEXT,
    title TEXT,
    payload_json TEXT,
    updated_at TEXT,
    PRIMARY KEY(profile, id)
);

CREATE TABLE IF NOT EXISTS photos (
    profile TEXT,
    id TEXT,
    payload_json TEXT,
    updated_at TEXT,
    PRIMARY KEY(profile, id)
);

CREATE TABLE IF NOT EXISTS checksums (
    profile TEXT,
    photo_id TEXT,
    algorithm TEXT,
    value TEXT,
    updated_at TEXT,
    PRIMARY KEY(profile, photo_id, algorithm)
);

CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    profile TEXT,
    command TEXT,
    state TEXT,
    payload_json TEXT,
    created_at TEXT,
    updated_at TEXT
);

CREATE TABLE IF NOT EXISTS job_items (
    job_id TEXT,
    item_key TEXT,
    state TEXT,
    payload_json TEXT,
    error_json TEXT,
    updated_at TEXT,
    PRIMARY KEY(job_id, item_key)
);
`
