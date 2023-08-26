CREATE TABLE cons (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    aliases TEXT [] NOT NULL DEFAULT '{}',
    location TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL
);
