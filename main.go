package main

import (
	"log"
	"net/http"
	"os"

	"mogost-tools/tools"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to release mode.
	gin.SetMode(gin.ReleaseMode)

	log.Println("Starting Mogost Toolkit server...")

	r := gin.Default()

	// Static file service.
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Main page route.
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Mogost Toolkit",
		})
	})

	// Tool 1: File comparison.
	fileCompare := r.Group("/api/file-compare")
	{
		fileCompare.POST("/upload", tools.HandleFileCompareUpload)
		fileCompare.GET("/compare", tools.HandleFileCompare)
	}

	// Tool 2: CSV viewer.
	csvViewer := r.Group("/api/csv")
	{
		csvViewer.POST("/upload", tools.HandleCSVUpload)
		csvViewer.GET("/view", tools.HandleCSVView)
	}

	// Tool 3: Archive extraction and trade comparison.
	archiveCompare := r.Group("/api/archive-compare")
	{
		archiveCompare.POST("/upload", tools.HandleArchiveUpload)
		archiveCompare.GET("/compare", tools.HandleArchiveCompare)
	}

	// Create required directories.
	createDirectories()

	log.Println("Mogost Toolkit server is running on :8080")
	log.Println("Open http://localhost:8080 in your browser")
	log.Fatal(r.Run(":8080"))
}

func createDirectories() {
	dirs := []string{
		"uploads",
		"uploads/file-compare",
		"uploads/csv",
		"uploads/archive-compare",
		"static",
		"templates",
		"temp",
	}

	log.Println("Creating required directories...")
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create directory %s: %v", dir, err)
		} else {
			log.Printf("Created directory: %s", dir)
		}
	}
}
