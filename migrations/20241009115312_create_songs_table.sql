-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE songs (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            group_name VARCHAR(255) NOT NULL,
            song_title VARCHAR(255) NOT NULL,
            release_date TIMESTAMP,
            text TEXT,
            link VARCHAR(255),
            verses JSONB
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE songs;
-- +goose StatementEnd
