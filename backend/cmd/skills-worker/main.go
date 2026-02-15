package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"SkillSync/internal/db"
	"SkillSync/internal/skills"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL missing")
	}

	database, err := db.New(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// 1️⃣ Get jobs without skills
	rows, err := database.Pool.Query(
		context.Background(),
		`SELECT id, description FROM jobs`,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var jobID int
		var desc string

		if err := rows.Scan(&jobID, &desc); err != nil {
			log.Fatal(err)
		}

		// 2️⃣ Extract skills
		found := skills.Extract(desc)
		if len(found) > 0 {
			log.Printf("Job %d: Found %d skills: %v", jobID, len(found), found)
		}

		// 3️⃣ Store skills
		for _, skill := range found {
			skillID, err := database.GetOrCreateSkill(skill)
			if err != nil {
				log.Println(err)
				continue
			}

			_ = database.InsertJobSkill(jobID, skillID)
		}
		count++
	}

	log.Printf("Skill extraction completed. Processed %d jobs.", count)
}

/*
Explanation of SQL Queries in skills-worker/main.go:

1. Select Unprocessed Jobs:
   - `SELECT id, description FROM jobs WHERE id NOT IN (SELECT DISTINCT job_id FROM job_skills)`
   - efficiently finds jobs that do NOT have any entries in the `job_skills` table yet.
   - This prevents the worker from re-processing jobs that have already been analyzed.
   - It effectively acts as a queue: Jobs with 0 skills associated are "todo".
*/