CREATE TABLE docs (
	id BIGSERIAL PRIMARY KEY,
	text text NOT NULL,
	url text NOT NULL
);

CREATE INDEX idx_docs_url ON docs (url);
