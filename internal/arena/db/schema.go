package db

// SchemaVersion is the on-disk schema revision. Bump this manually on every
// schema change so existing user databases are dropped and recreated on the
// next Open instead of failing later with errors against an outdated layout.
const SchemaVersion = "1"

const schema = `
CREATE TABLE IF NOT EXISTS schema_version (
    version TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS puzzles (
    pretty_id      TEXT PRIMARY KEY,
    leaderboard_id TEXT NOT NULL,
    fetched_at     INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS players (
    leaderboard_id TEXT NOT NULL,
    nickname       TEXT NOT NULL,
    agent_id       INTEGER NOT NULL,
    fetched_at     INTEGER NOT NULL,
    PRIMARY KEY (leaderboard_id, nickname)
);
`
