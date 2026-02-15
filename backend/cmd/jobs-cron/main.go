package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"

	"github.com/joho/godotenv"

	"SkillSync/internal/adzuna"
	"SkillSync/internal/db"
)

func makeJobHash(title, company, location string) string {
	key := title + "|" + company + "|" + location
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func main() {
	_ = godotenv.Load()

	appID := os.Getenv("ADZUNA_APP_ID")
	appKey := os.Getenv("ADZUNA_APP_KEY")
	dbURL := os.Getenv("DATABASE_URL")

	if appID == "" || appKey == "" || dbURL == "" {
		log.Fatal("Missing required environment variables")
	}

	// 1️⃣ Connect DB
	database, err := db.New(dbURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer database.Close()

	// 2️⃣ Create Adzuna client
	client := adzuna.NewClient(appID, appKey)

	// 3️⃣ Fetch jobs
	jobs, err := client.FetchJobs("software", "india", 3)
	if err != nil {
		log.Fatalf("Failed to fetch jobs: %v", err)
	}

	log.Printf("Fetched %d jobs", len(jobs))

	// 4️⃣ Insert jobs with hashing
	inserted := 0

	for _, job := range jobs {
		job.JobHash = makeJobHash(job.Title, job.Company, job.Location)

		err := database.InsertJob(job)
		if err != nil {
			log.Printf("Insert failed (%s): %v", job.Title, err)
			continue
		}
		inserted++
	}

	log.Printf("Inserted %d new jobs", inserted)
}




