package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"SkillSync/internal/db"
	"SkillSync/internal/resume"
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

	// 1️⃣ Create (or get) user
	userID, err := database.CreateUser("test@example.com")
	if err != nil {
		log.Fatal(err)
	}

	// 2️⃣ Fake resume text (for now)
	rawResume := `
		Backend developer with experience in Go, PostgreSQL,
		Docker, Linux, and AWS. Worked on REST APIs.
	`

	// 3️⃣ Parse resume
	text := resume.Parse(rawResume)

	// 4️⃣ Extract skills
	found := skills.Extract(text)

	// 5️⃣ Store user skills
	for _, skill := range found {
		skillID, err := database.GetOrCreateSkill(skill)
		if err != nil {
			log.Println(err)
			continue
		}
		_ = database.InsertUserSkill(userID, skillID)
	}

	log.Println("User resume processed successfully")
}
