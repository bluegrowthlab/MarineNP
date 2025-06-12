/*
 * MarineNP Main Application
 * Purpose: Main entry point for the MarineNP API server
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file initializes and runs the MarineNP API server, which provides
 * access to marine natural products data through a RESTful API.
 */

package main

import (
	"fmt"
	"log"
	"time"

	"marinenp/config"
	"marinenp/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Configuration
	// Load application configuration from environment variables
	cfg := config.LoadConfig()

	// Database Initialization
	// Set up database connection with appropriate configuration
	var db *gorm.DB
	var err error
	
	if cfg.Database.Type == "sqlite" {
		db, err = gorm.Open(sqlite.Open(cfg.Database.GetDSN()), &gorm.Config{
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
			PrepareStmt: true,
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	} else {
		log.Fatal("Only SQLite database type is supported")
	}
	
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Database Configuration
	// Configure SQLite connection pool and timestamp handling
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Connection Pool Settings
	// Optimize database connection pool for performance
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Handler Setup
	// Initialize database connection in request handlers
	handlers.SetDB(db)

	// Router Setup
	// Initialize Gin router with CORS configuration
	r := gin.Default()

	// CORS Configuration
	// Enable Cross-Origin Resource Sharing with appropriate headers
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", cfg.API.CorsAllowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// API Routes
	// Define all API endpoints under /api/v1/
	api := r.Group("/api/v1")
	{
		// Statistics Endpoints
		api.GET("/statistics", handlers.GetStatistics)

		// Molecules Endpoints
		// Endpoints for accessing and analyzing molecular data
		api.GET("/molecules/:identifier", handlers.GetMoleculeByID)
		api.GET("/molecules/search", handlers.SearchMolecules)
		api.GET("/molecules/properties/ranges", handlers.GetPropertyRanges)
		api.GET("/molecules/export", handlers.ExportMolecules)
		api.GET("/molecules/analyze", handlers.AnalyzeMolecules)

		// Organisms Endpoints
		// Endpoints for accessing organism data and their associated molecules
		api.GET("/organisms", handlers.GetOrganisms)
		api.GET("/organisms/:id", handlers.GetOrganismByID)
		api.GET("/organisms/:id/molecules", handlers.GetMoleculesByOrganism)
		api.GET("/organisms/autocomplete", handlers.GetOrganismsAutocomplete)

		// Collections Endpoints
		// Endpoints for accessing collection data and their molecules
		api.GET("/collections", handlers.GetCollections)
		api.GET("/collections/:id", handlers.GetCollectionByID)
		api.GET("/collections/:id/molecules", handlers.GetMoleculesByCollection)

		// Citations Endpoints
		// Endpoints for accessing literature citations
		api.GET("/citations", handlers.GetCitations)
		api.GET("/citations/:id", handlers.GetCitationByID)

		// Geographic Locations Endpoints
		// Endpoints for accessing location data and associated molecules
		api.GET("/locations", handlers.GetLocations)
		api.GET("/locations/:id", handlers.GetLocationByID)
		api.GET("/locations/:id/molecules", handlers.GetMoleculesByLocation)

		// OBIS Integration Endpoints
		// Endpoints for accessing Ocean Biogeographic Information System data
		api.GET("/obis/locations", handlers.GetOBISLocations)
	}

	// Static File Serving
	// Serve static files for the web interface
	r.StaticFile("/", "./public/index.html")
	r.Static("/public", "./public/public")
	r.Static("/pages", "./public/pages")

	// Server Startup
	// Start the HTTP server on the configured port
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("\n🚀 Server is running! Visit http://localhost%s\n\n", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
} 