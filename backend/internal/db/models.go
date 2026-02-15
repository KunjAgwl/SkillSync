package db

import "time"

type Job struct {
	ID          int
	JobHash     string
	Title       string
	Company     string
	Location    string
	Description string
	ApplyURL    string
	Source      string
	CreatedAt   time.Time
}

type User struct {
	ID           int
	Email        string
	PasswordHash string
	ResumeText   string
	CreatedAt    time.Time
}

type Skill struct {
	ID   int
	Name string
}
