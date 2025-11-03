package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin mode (release mode in production)
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default() // Includes Logger and Recovery middleware

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
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

	// API routes group (will be populated as we add handlers)
	_ = r.Group("/api")
	// TODO: Uncomment and use when adding handlers:
	// api := r.Group("/api")
	// {
	//   api.POST("/upload", handlers.HandleUpload)
	//   api.GET("/session/:id", handlers.HandleGetSession)
	//   api.POST("/session/:id/answers", handlers.HandleSubmitAnswers)
	//   api.GET("/session/:id/next", handlers.HandleGetNextQuestion)
	//   api.POST("/session/:id/ai/questions", handlers.HandleGenerateQuestions)
	//   api.POST("/session/:id/generate", handlers.HandleGenerateDocument)
	// }

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
