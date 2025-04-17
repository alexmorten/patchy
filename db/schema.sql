CREATE TABLE docs (
	id BIGSERIAL PRIMARY KEY,
	text text NOT NULL,
	url text NOT NULL,
	message_id text NOT NULL UNIQUE
);

CREATE INDEX idx_docs_url ON docs (url);
CREATE INDEX idx_docs_message_id ON docs (message_id);
