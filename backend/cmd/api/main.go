package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"SkillSync/internal/auth"
	"SkillSync/internal/db"
	"SkillSync/internal/embedding"
	"SkillSync/internal/match"
	"SkillSync/internal/resume"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

// Middleware to enable CORS
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	_ = godotenv.Load()

	// 1. Database Connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	database, err := db.New(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// 2. Initialize Gemini Client
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set")
	}
	geminiClient := embedding.NewGemini(geminiKey)

	// 3. HTTP Handlers

	// AUTH - REGISTER
	http.HandleFunc("/register", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		id, err := database.CreateUser(req.Email, hash)
		if err != nil {
			log.Println(err)
			http.Error(w, "User already exists or database error", http.StatusConflict)
			return
		}

		// Generate JWT for the new user
		token, err := auth.GenerateJWT(id, req.Email)
		if err != nil {
			log.Printf("Token generation error: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(LoginResponse{
			Token: token,
			User: struct {
				ID    int    `json:"id"`
				Email string `json:"email"`
			}{
				ID:    id,
				Email: req.Email,
			},
		})
	}))

	// RESUME - UPLOAD
	http.HandleFunc("/upload-resume", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received /upload-resume request")
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			log.Println("Invalid method for /upload-resume:", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Validate JWT Token
		authHeader := r.Header.Get("Authorization")
		log.Println("Authorization Header:", authHeader)
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		log.Printf("Token valid for user ID: %d", claims.UserID)

		// Parse multipart form (max 10MB)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Handle Resume File
		file, header, err := r.FormFile("resume")
		if err != nil {
			http.Error(w, "Resume file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		// Extract text from resume
		resumeText, err := resume.Parse(fileBytes, header.Filename)
		if err != nil {
			log.Printf("Resume parse error: %v", err)
			http.Error(w, "Failed to parse resume file", http.StatusBadRequest)
			return
		}

		if len(resumeText) < 50 {
			http.Error(w, "Resume text too short or unreadable", http.StatusBadRequest)
			return
		}

		// Update DB with Resume Text
		err = database.UpdateResume(claims.UserID, resumeText)
		if err != nil {
			log.Println(err)
			http.Error(w, "Database update error", http.StatusInternalServerError)
			return
		}

		// --- GENERATE & SAVE EMBEDDING IMMEDIATELY ---
		log.Printf("Generating embedding for user %d...", claims.UserID)
		embeddingVec, err := geminiClient.Embed(resumeText)
		if err != nil {
			log.Printf("Embedding generation failed for user %d: %v", claims.UserID, err)
			http.Error(w, "Resume uploaded but failed to generate embedding", http.StatusInternalServerError)
			return
		}

		err = database.UpdateUserEmbedding(claims.UserID, embeddingVec)
		if err != nil {
			log.Printf("Embedding save failed for user %d: %v", claims.UserID, err)
			http.Error(w, "Resume uploaded but failed to save embedding", http.StatusInternalServerError)
			return
		}
		log.Printf("Embedding saved for user %d", claims.UserID)
		// ---------------------------------------------

		json.NewEncoder(w).Encode(map[string]string{
			"message": "Resume uploaded and processed successfully",
		})
	}))

	// AUTH - LOGIN
	http.HandleFunc("/login", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received /login request")
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			log.Println("Invalid method for /login:", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Login request body decode failed: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		log.Printf("Login attempt for email: %s", req.Email)

		user, err := database.GetUserByEmail(req.Email)
		if err != nil {
			log.Printf("User not found or DB error: %v", err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
			log.Println("Password mismatch")
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		token, err := auth.GenerateJWT(user.ID, user.Email)
		if err != nil {
			log.Printf("Token generation failed: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		log.Printf("Login successful for user %d", user.ID)

		json.NewEncoder(w).Encode(LoginResponse{
			Token: token,
			User: struct {
				ID    int    `json:"id"`
				Email string `json:"email"`
			}{
				ID:    user.ID,
				Email: user.Email,
			},
		})
	}))

	http.HandleFunc("/matches", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		// Set JSON content type
		w.Header().Set("Content-Type", "application/json")

		// Parse user_id from query params
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
			return
		}

		// Call the matching logic
		matches, err := match.MatchJobs(database, userID)
		if err != nil {
			log.Printf("Match error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Encode response
		if err := json.NewEncoder(w).Encode(matches); err != nil {
			log.Printf("JSON encode error: %v", err)
		}
	}))

	// 4. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
