package db

const schema = `
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
