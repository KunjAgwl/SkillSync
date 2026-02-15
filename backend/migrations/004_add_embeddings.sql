ALTER TABLE jobs
ADD COLUMN embedding VECTOR(768);

ALTER TABLE users
ADD COLUMN embedding VECTOR(768);



-- CREATE INDEX jobs_embedding_idx
-- ON jobs
-- USING ivfflat (embedding vector_cosine_ops);

-- CREATE INDEX users_embedding_idx
-- ON users
-- USING ivfflat (embedding vector_cosine_ops);


-- ALTER TABLE users
-- ADD COLUMN resume_text TEXT;
