/*
 * MarineNP Citations Handlers
 * Purpose: HTTP handlers for managing and retrieving citation data
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file provides endpoints for accessing and managing literature citations
 * associated with marine natural products in the MarineNP database.
 */

package handlers

import (
	"marinenp/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCitations handles GET /api/v1/citations
func GetCitations(c *gin.Context) {
	params := ParseQueryParams(c)
	var citations []models.Citation
	var total int64

	// Build query
	query := db.Model(&models.Citation{})

	// Apply search if provided
	if params.Search != "" {
		query = query.Where("LOWER(title) LIKE ? OR LOWER(authors) LIKE ? OR LOWER(doi) LIKE ?",
			"%"+strings.ToLower(params.Search)+"%", "%"+strings.ToLower(params.Search)+"%", "%"+strings.ToLower(params.Search)+"%")
	}

	// Get total count
	query.Count(&total)

	// Apply ordering
	if params.OrderByString != "" {
		order := params.OrderByString
		if params.OrderDir == "desc" {
			order += " DESC"
		}
		query = query.Order("citations.id ASC").Order(order)
	} else {
		query = query.Order("citations.id ASC")
	}

	// Apply pagination
	offset := (params.PageNumber - 1) * params.PerPageNumber
	query = query.Offset(offset).Limit(params.PerPageNumber)

	// Execute query
	result := query.Find(&citations)

	if result.Error != nil {
		ErrorResponse(c, 500, "Failed to fetch citations")
		return
	}

	// Marshal the response using our custom marshaler
	jsonData, err := models.MarshalToJSON(gin.H{
		"citations": citations,
		"total":    total,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

// GetCitationByID handles GET /api/v1/citations/:id
func GetCitationByID(c *gin.Context) {
	id := c.Param("id")
	var citation models.Citation

	result := db.First(&citation, id)

	if result.Error != nil {
		ErrorResponse(c, 404, "Citation not found")
		return
	}

	SuccessResponse(c, citation)
} 