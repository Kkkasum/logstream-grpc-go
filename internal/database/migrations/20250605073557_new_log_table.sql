-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS logs (
    id SERIAL PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    source_id VARCHAR(255) NOT NULL,
    lvl SMALLINT NOT NULL,
    message VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logs;
-- +goose StatementEnd
