package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"SkillSync/internal/db"
	"SkillSync/internal/embedding"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	apiKey := os.Getenv("GEMINI_API_KEY")

	if dbURL == "" || apiKey == "" {
		log.Fatal("Missing DATABASE_URL or EMBEDDING_API_KEY")
	}

	// 1️⃣ DB
	database, err := db.New(dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// 2️⃣ Embedding client
	embedder := embedding.NewGemini(os.Getenv("GEMINI_API_KEY"))

	// 3️⃣ Fetch jobs without embeddings
	jobs, err := database.GetJobsWithoutEmbeddings(50)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d jobs without embeddings", len(jobs))

	// 4️⃣ Generate & store embeddings
	for _, job := range jobs {
		vec, err := embedder.Embed(job.Text)
		if err != nil {
			log.Println("Embedding failed:", err)
			continue
		}

		err = database.UpdateJobEmbedding(job.ID, vec)
		if err != nil {
			log.Println("DB update failed:", err)
			continue
		}
	}

	log.Println("Embedding worker completed")

	// ===== USER EMBEDDINGS =====

	users, err := database.GetUsersWithoutEmbeddings(20)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d users without embeddings", len(users))

	for _, user := range users {
		vec, err := embedder.Embed(user.Text)
		if err != nil {
			log.Println("User embedding failed:", err)
			continue
		}

		err = database.UpdateUserEmbedding(user.ID, vec)
		if err != nil {
			log.Println("User embedding DB update failed:", err)
			continue
		}
	}

}
