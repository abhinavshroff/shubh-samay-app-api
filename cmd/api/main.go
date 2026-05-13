package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/shubh-samay/api/internal/config"
	"github.com/shubh-samay/api/internal/db"
	"github.com/shubh-samay/api/internal/handlers"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(pool); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	r := gin.Default()
	r.Use(corsMiddleware())

	v1 := r.Group("/v1")
	{
		v1.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
		v1.GET("/panchang", handlers.GetPanchang)
		v1.GET("/festivals", handlers.GetFestivals)
		v1.GET("/lunar-days", handlers.GetLunarDays)
		v1.GET("/muhurat", handlers.FindMuhurat)
		v1.GET("/config", handlers.GetConfig(pool))
		v1.POST("/devices", handlers.RegisterDevice(pool))

		admin := v1.Group("/admin")
		admin.Use(adminAuth(cfg.AdminToken))
		{
			admin.PATCH("/flags/:key", handlers.UpdateFlag(pool))
			admin.GET("/flags", handlers.ListFlags(pool))
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Shubh Samay API listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func adminAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Admin-Token") != token {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
