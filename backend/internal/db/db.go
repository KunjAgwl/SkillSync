package db

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the pgx pool
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new database connection pool
func New(databaseURL string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &DB{Pool: pool}, nil
}

// Close shuts down the pool
func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) InsertJob(job Job) error {
	query := `
		INSERT INTO jobs
		(job_hash, title, company, location, description, apply_url, source)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (job_hash) DO NOTHING
	`

	_, err := db.Pool.Exec(
		context.Background(),
		query,
		job.JobHash,
		job.Title,
		job.Company,
		job.Location,
		job.Description,
		job.ApplyURL,
		job.Source,
	)

	return err
}

func (db *DB) GetOrCreateSkill(name string) (int, error) {
	var id int

	query := `
		INSERT INTO skills (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`

	err := db.Pool.QueryRow(
		context.Background(),
		query,
		name,
	).Scan(&id)

	return id, err
}

func (db *DB) InsertJobSkill(jobID, skillID int) error {
	query := `
		INSERT INTO job_skills (job_id, skill_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	_, err := db.Pool.Exec(
		context.Background(),
		query,
		jobID,
		skillID,
	)

	return err
}

func (db *DB) CreateUser(email, passwordHash string) (int, error) {
	var id int

	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	err := db.Pool.QueryRow(
		context.Background(),
		query,
		email,
		passwordHash,
	).Scan(&id)

	return id, err
}

func (db *DB) UpdateResume(userID int, resumeText string) error {
	query := `
		UPDATE users
		SET resume_text = $2
		WHERE id = $1
	`
	_, err := db.Pool.Exec(context.Background(), query, userID, resumeText)
	return err
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	var user User
	query := `
		SELECT id, email, password_hash
		FROM users
		WHERE email = $1
	`
	err := db.Pool.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) InsertUserSkill(userID, skillID int) error {
	query := `
		INSERT INTO user_skills (user_id, skill_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	_, err := db.Pool.Exec(
		context.Background(),
		query,
		userID,
		skillID,
	)

	return err
}

func (db *DB) GetJobsWithoutEmbeddings(limit int) ([]struct {
	ID   int
	Text string
}, error) {

	rows, err := db.Pool.Query(context.Background(), `
		SELECT
			id,
			title || ' ' || description AS text
		FROM jobs
		WHERE embedding IS NULL
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []struct {
		ID   int
		Text string
	}

	for rows.Next() {
		var j struct {
			ID   int
			Text string
		}
		if err := rows.Scan(&j.ID, &j.Text); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

func (db *DB) formatVector(vec []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, v := range vec {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

func (db *DB) UpdateJobEmbedding(jobID int, vec []float32) error {
	vecStr := db.formatVector(vec)

	_, err := db.Pool.Exec(
		context.Background(),
		`
		UPDATE jobs
		SET embedding = $2::vector
		WHERE id = $1
		`,
		jobID,
		vecStr,
	)

	return err
}

func (db *DB) GetUsersWithoutEmbeddings(limit int) ([]struct {
	ID   int
	Text string
}, error) {

	rows, err := db.Pool.Query(context.Background(), `
		SELECT id, resume_text
		FROM users
		WHERE embedding IS NULL
		  AND resume_text IS NOT NULL
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []struct {
		ID   int
		Text string
	}

	for rows.Next() {
		var u struct {
			ID   int
			Text string
		}
		if err := rows.Scan(&u.ID, &u.Text); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (db *DB) UpdateUserEmbedding(userID int, vec []float32) error {
	vecStr := db.formatVector(vec)

	_, err := db.Pool.Exec(
		context.Background(),
		`
		UPDATE users
		SET embedding = $2::vector
		WHERE id = $1
		`,
		userID,
		vecStr,
	)

	return err
}

/*
Explanation of SQL Queries in db.go:

1. InsertJob:
   - Inserts a new job record.
   - uses `ON CONFLICT (job_hash) DO NOTHING` to prevent duplicate jobs from being inserted multiple times.

2. GetOrCreateSkill:
   - Tries to insert a skill name.
   - If it exists (`ON CONFLICT`), it updates the name (idempotent) and returns the ID.

3. InsertJobSkill / InsertUserSkill:
   - Links a job or user to a skill in the many-to-many junction tables.
   - Uses `DO NOTHING` on conflict to avoid duplicate pairs.

4. CreateUser:
   - Inserts a user email.
   - Upserts on conflict to ensure we just return the existing ID if they are already there.

5. GetJobsWithoutEmbeddings / GetUsersWithoutEmbeddings:
   - Finds rows where the `embedding` column IS NULL.
   - Used by background workers to find content that needs processing.
   - `GetUsersWithoutEmbeddings` also checks `resume_text IS NOT NULL` because we can't embed nothing.

6. UpdateJobEmbedding / UpdateUserEmbedding:
   - Saves the generated vector back to the database.
   - Casts the string representation to `::vector` for the pgvector extension.
*/
