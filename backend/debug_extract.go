package main

import (
	"SkillSync/internal/skills"
	"fmt"
)

func main() {
	text := "We are looking for a React Native developer with Node.js and Python experience. Knowledge of AWS and Docker is a plus."
	extracted := skills.Extract(text)
	fmt.Printf("Input: %s\n", text)
	fmt.Printf("Extracted: %v\n", extracted)
}
