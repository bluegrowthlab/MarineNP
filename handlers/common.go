/*
 * MarineNP Common Handlers
 * Purpose: Common HTTP request handlers and response utilities
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file contains shared functionality for handling HTTP requests and responses,
 * including pagination, error handling, and common query parameter parsing.
 */

package handlers

import (
	"marinenp/models"
	"net/http"
	"strconv"

	"marinenp/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB

// SetDB initializes the database connection for all handlers
func SetDB(database *gorm.DB) {
	db = database
}

// Response represents the standard AMIS API response format
type Response struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

// PaginatedResponse represents the standard paginated data structure
type PaginatedResponse struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
}

// QueryParams represents common query parameters for API requests
type QueryParams struct {
	PageNumber    int
	PerPageNumber int
	OrderByString string
	OrderDir      string
	Search        string
}

// ParseQueryParams extracts and validates common query parameters from the request
func ParseQueryParams(c *gin.Context) QueryParams {
	// Parse page number with fallback to pageNumber parameter
	pageNumber, _ := strconv.Atoi(c.DefaultQuery("page", c.DefaultQuery("pageNumber", "1")))
	// Parse items per page with fallback to perPageNumber parameter
	perPageNumber, _ := strconv.Atoi(c.DefaultQuery("perPage", c.DefaultQuery("perPageNumber", "10")))
	
	// Validate and set default values
	if perPageNumber <= 0 {
		perPageNumber = 10
	}
	if pageNumber <= 0 {
		pageNumber = 1
	}

	return QueryParams{
		PageNumber:    pageNumber,
		PerPageNumber: perPageNumber,
		OrderByString: c.DefaultQuery("orderByString", ""),
		OrderDir:      c.DefaultQuery("orderDir", "asc"),
		Search:        c.DefaultQuery("query", ""),
	}
}

// SuccessResponse sends a standardized successful response
func SuccessResponse(c *gin.Context, data interface{}) {
	// Marshal response using custom JSON marshaler
	jsonData, err := models.MarshalToJSON(gin.H{
		"status": 0,
		"data": data,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

// ErrorResponse sends a standardized error response
func ErrorResponse(c *gin.Context, status int, msg string) {
	c.JSON(http.StatusOK, Response{
		Status: status,
		Msg:    msg,
		Data:   nil,
	})
}

// PaginatedSuccessResponse sends a standardized paginated response
func PaginatedSuccessResponse(c *gin.Context, data interface{}, total int64, page int) {
	// Marshal response using custom JSON marshaler
	jsonData, err := models.MarshalToJSON(gin.H{
		"data":  data,
		"total": total,
		"page":  page,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

// GetStatistics retrieves and returns system-wide statistics
func GetStatistics(c *gin.Context) {
	var stats struct {
		TotalMolecules int64  `json:"total_molecules"`
		TotalOrganisms int64  `json:"total_organisms"`
		Version        string `json:"version"`
		LastUpdate     string `json:"last_update"`
	}

	// Count marine molecules
	db.Model(&models.Molecule{}).Where("is_marine = TRUE").Count(&stats.TotalMolecules)

	// Count marine organisms
	db.Model(&models.Organism{}).Where("is_marine = TRUE").Count(&stats.TotalOrganisms)

	// Get version and last update information
	cfg := config.LoadConfig()
	stats.Version = cfg.Version
	stats.LastUpdate = cfg.LastUpdate

	SuccessResponse(c, stats)
} 