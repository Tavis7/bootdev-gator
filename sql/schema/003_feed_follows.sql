-- +goose Up
CREATE TABLE feed_follows(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id),
    feed_id UUID NOT NULL REFERENCES feeds ON DELETE CASCADE,
    CONSTRAINT fk_feed_id FOREIGN KEY (feed_id) REFERENCES feeds(id),
    UNIQUE(feed_id, user_id)
);

-- +goose Down
DROP TABLE feed_follows;
