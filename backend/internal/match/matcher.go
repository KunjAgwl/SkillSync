package match

import (
	"context"
	"sort"

	"SkillSync/internal/db"
)

// JobMatch represents a ranked job match for a user
type JobMatch struct {
	JobID       int    `json:"job_id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Description string `json:"description"`
	ApplyURL    string `json:"apply_url"`

	SkillScore float64 `json:"skill_score"`
	EmbedScore float64 `json:"embed_score"`
	MatchScore float64 `json:"match_score"`

	MissingSkills []string `json:"missing_skills"`
}

// getEmbeddingScores fetches semantic similarity scores for jobs vs a user
func getEmbeddingScores(database *db.DB, userID int) (map[int]float64, error) {
	rows, err := database.Pool.Query(context.Background(), `
		SELECT
			j.id,
			1 - (j.embedding <-> u.embedding) AS similarity
		FROM jobs j
		JOIN users u ON u.id = $1
		WHERE j.embedding IS NOT NULL
		  AND u.embedding IS NOT NULL
		ORDER BY similarity DESC
		LIMIT 200
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := make(map[int]float64)

	for rows.Next() {
		var jobID int
		var sim float64
		if err := rows.Scan(&jobID, &sim); err != nil {
			return nil, err
		}

		// Clamp defensively
		if sim < 0 {
			sim = 0
		}
		if sim > 1 {
			sim = 1
		}

		scores[jobID] = sim
	}

	return scores, nil
}

// MatchJobs returns hybrid-ranked jobs for a user
func MatchJobs(database *db.DB, userID int) ([]JobMatch, error) {
	ctx := context.Background()

	// 1️⃣ Fetch embedding similarity scores
	embeddingScores, err := getEmbeddingScores(database, userID)
	if err != nil {
		return nil, err
	}

	// 2️⃣ Fetch skill-based matches
	rows, err := database.Pool.Query(ctx, `
		SELECT
			j.id,
			j.title,
			j.company,
			j.location,
			j.description,
			j.apply_url,
			CASE
				WHEN COUNT(js.skill_id) = 0 THEN 0
				ELSE COUNT(us.skill_id)::FLOAT / COUNT(js.skill_id)
			END AS skill_score
		FROM jobs j
		LEFT JOIN job_skills js ON j.id = js.job_id
		LEFT JOIN user_skills us
			ON js.skill_id = us.skill_id
			AND us.user_id = $1
		GROUP BY j.id
		ORDER BY skill_score DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []JobMatch{}

	for rows.Next() {
		var jm JobMatch

		if err := rows.Scan(
			&jm.JobID,
			&jm.Title,
			&jm.Company,
			&jm.Location,
			&jm.Description,
			&jm.ApplyURL,
			&jm.SkillScore,
		); err != nil {
			return nil, err
		}

		// 3️⃣ Blend embedding score
		embedScore := embeddingScores[jm.JobID]
		jm.EmbedScore = embedScore

		// Hybrid score (tunable)
		jm.MatchScore = 0.7*jm.SkillScore + 0.3*embedScore

		// 4️⃣ Fetch missing skills
		missRows, err := database.Pool.Query(ctx, `
			SELECT s.name
			FROM skills s
			JOIN job_skills js ON s.id = js.skill_id
			WHERE js.job_id = $1
			  AND s.id NOT IN (
				  SELECT skill_id
				  FROM user_skills
				  WHERE user_id = $2
			  )
		`, jm.JobID, userID)
		if err != nil {
			return nil, err
		}

		for missRows.Next() {
			var skill string
			_ = missRows.Scan(&skill)
			jm.MissingSkills = append(jm.MissingSkills, skill)
		}
		missRows.Close()

		results = append(results, jm)
	}

	// 5️⃣ Sort by final hybrid score
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchScore > results[j].MatchScore
	})

	return results, nil
}

/*
Explanation of SQL Queries in matcher.go:

1. getEmbeddingScores (Vector Similarity):
   - Computes Semantic Similarity using `1 - (j.embedding <-> u.embedding)`.
   - The `<->` operator calculates Cosine Distance (lower is better, 0 is identical).
   - We subtract from 1 to get Similarity (higher is better, 1 is identical).
   - Filters out rows with NULL embeddings.

2. MatchJobs (Skill & Hybrid Matching):
   - Calculates a `skill_score` based on the overlap of user skills vs job skills.
   - Formula: `COUNT(user_matched_skills) / COUNT(total_job_skills)`.
   - Uses `LEFT JOIN` on `job_skills` so that jobs with 0 skills are NOT excluded.
   - Handles division by zero (if a job has 0 skills extracted yet) by defaulting score to 0.
   - Fetches the top 100 candidates based on skill overlap.
   - This list is then re-ranked in Go using the weighted hybrid score:
     `FinalScore = 0.7 * SkillScore + 0.3 * EmbeddingScore`

3. Fetch Missing Skills:
   - For each matched job, finds the specific skill names that the job requires but the user lacks.
   - `WHERE s.id NOT IN (SELECT skill_id FROM user_skills ...)`
*/
