package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/you/lexsy-mvp/server/handlers"
	"github.com/you/lexsy-mvp/server/session"
)

func main() {
	// Set Gin mode (release mode in production)
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default() // Includes Logger and Recovery middleware

	// Initialize session store
	store := session.NewStore()

	// CORS configuration
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173,http://localhost:3000"
	}
	origins := strings.Split(allowedOrigins, ",")
	// Trim whitespace from each origin
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	log.Printf("CORS allowed origins: %v", origins)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		api.POST("/upload", handlers.HandleUpload(store))
		api.GET("/session/:id", handlers.HandleGetSession(store))
		api.POST("/session/:id/answers", handlers.HandleSubmitAnswers(store))
		api.GET("/session/:id/next", handlers.HandleGetNextQuestion(store))
		api.POST("/session/:id/ai/questions", handlers.HandleGenerateQuestions(store))
		api.POST("/session/:id/generate", handlers.HandleGenerateDocument(store))
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
