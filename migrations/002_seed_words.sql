-- A small word list so the game works out of the box.
-- Students can extend this; the game picks one per day, deterministically.
INSERT INTO words (word) VALUES
    ('CRANE'),('SLATE'),('AUDIO'),('ROAST'),('PLUMB'),
    ('GHOST'),('MIRTH'),('QUILT'),('VOWEL'),('BRAVE'),
    ('CHALK'),('DWELL'),('FROST'),('GRIME'),('HOUND'),
    ('IVORY'),('JOKER'),('KNEEL'),('LUMEN'),('MANGO'),
    ('NUDGE'),('OPERA'),('PRIZE'),('QUERY'),('RIVET'),
    ('SHINE'),('TWIST'),('UNITY'),('VAPOR'),('WHARF'),
    ('EERIE'),('ABBEY'),('LLAMA'),('SASSY'),('TEPEE')
ON CONFLICT DO NOTHING;
