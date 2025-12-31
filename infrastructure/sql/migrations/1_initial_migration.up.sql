CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL CHECK(name <> ''),
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(255) NOT NULL,
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMP
);

CREATE TABLE session(
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE song(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE playlist(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE playlist_song(
    playlist_id INTEGER REFERENCES playlist(id) ON DELETE CASCADE,
    song_id INTEGER REFERENCES song(id) ON DELETE CASCADE,
    PRIMARY KEY (playlist_id, song_id)
);

CREATE INDEX song_name_idx ON song(name);
CREATE INDEX playlist_name_idx ON playlist(name);

CREATE TABLE artist(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE song_artist(
    song_id INTEGER REFERENCES song(id) ON DELETE CASCADE,
    artist_id INTEGER REFERENCES artist(id) ON DELETE CASCADE,
    PRIMARY KEY (song_id, artist_id)
);

CREATE INDEX artist_name_idx ON artist(name);
CREATE INDEX song_artist_idx ON song_artist(song_id, artist_id);
CREATE INDEX song_artist_song_idx ON song_artist(song_id);

CREATE TABLE playlist_artist(
    playlist_id INTEGER REFERENCES playlist(id) ON DELETE CASCADE,
    artist_id INTEGER REFERENCES artist(id) ON DELETE CASCADE,
    PRIMARY KEY (playlist_id, artist_id)
);

CREATE INDEX playlist_artist_idx ON playlist_artist(playlist_id, artist_id);
