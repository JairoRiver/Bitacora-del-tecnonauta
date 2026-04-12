CREATE TYPE block_type AS ENUM ('p', 'c', 'i', 't');

CREATE TABLE posts (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title      TEXT        NOT NULL,
    subtitle   TEXT        NOT NULL DEFAULT '',
    date       DATE        NOT NULL,
    hero_image TEXT        NOT NULL DEFAULT '',
    slug       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE categories (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE
);

CREATE TABLE post_categories (
    post_id     UUID NOT NULL REFERENCES posts(id)      ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, category_id)
);

-- conf usa DEFAULT 'null'::jsonb para evitar NULLs SQL:
-- bloques sin configuración (ImageBlock) almacenan JSON null.
CREATE TABLE content_blocks (
    id       UUID       PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id  UUID       NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    position INTEGER    NOT NULL,
    type     block_type NOT NULL,
    value    JSONB      NOT NULL,
    conf     JSONB      NOT NULL DEFAULT 'null'::jsonb,
    UNIQUE (post_id, position)
);

CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
