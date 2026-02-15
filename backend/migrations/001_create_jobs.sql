CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,

    job_hash TEXT NOT NULL UNIQUE,

    title TEXT NOT NULL,
    company TEXT NOT NULL,
    location TEXT NOT NULL,

    description TEXT NOT NULL,
    apply_url TEXT NOT NULL,

    source TEXT NOT NULL DEFAULT 'adzuna',

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
