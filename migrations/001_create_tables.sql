CREATE TABLE IF NOT EXISTS words (
    word TEXT PRIMARY KEY CHECK (length(word) = 5)
);

CREATE TABLE IF NOT EXISTS results (
    id      SERIAL PRIMARY KEY,
    day     DATE    NOT NULL,
    player  TEXT    NOT NULL,
    guesses INT     NOT NULL,
    won     BOOLEAN NOT NULL,
    UNIQUE (day, player)
);

CREATE INDEX IF NOT EXISTS results_day ON results (day);
