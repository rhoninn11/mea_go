-- Ludzie z miasta i ich problemy
CREATE TABLE townsfolk (
    id       SERIAL PRIMARY KEY,
    name     TEXT NOT NULL,
    sorrow   TEXT NOT NULL   -- każdy ma swój smutek: zaginiony narzeczony, chora siostra...
);

-- To jest pamięć świata, some hint that allow you to peak deepe
CREATE TABLE places (
    id     SERIAL PRIMARY KEY,
    name   TEXT NOT NULL,
    about   TEXT NOT NULL
);

-- Ogromne istoty zamieszkujące ten świat, ludzie zwą ich zesłańcami.
CREATE TABLE colosses (
    id            SERIAL PRIMARY KEY,
    place_id      INT NOT NULL REFERENCES places(id),
    size          TEXT NOT NULL CHECK (size IN ('duży', 'ogromny', 'jak góra')),
    attacks_first BOOLEAN NOT NULL DEFAULT false,  -- większość nie zaatakuje pierwsza
    alive         BOOLEAN NOT NULL DEFAULT true
);


CREATE TABLE world_secrets (
    id     SERIAL PRIMARY KEY,
    place_ref INT NOT NULL REFERENCES places(id),
    mystery   TEXT NOT NULL
);