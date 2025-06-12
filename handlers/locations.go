/*
 * MarineNP Locations Handlers
 * Purpose: HTTP handlers for managing and retrieving geographic location data
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file provides endpoints for accessing and managing geographic locations
 * associated with marine natural products in the MarineNP database.
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"marinenp/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetLocations handles GET /api/v1/locations
func GetLocations(c *gin.Context) {
	params := ParseQueryParams(c)
	var locations []models.GeoLocation
	var total int64

	// Build query
	query := db.Model(&models.GeoLocation{})

	// Apply search if provided
	if params.Search != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(country) LIKE ? OR LOWER(region) LIKE ?",
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
		query = query.Order("geo_locations.id ASC").Order(order)
	} else {
		query = query.Order("geo_locations.id ASC")
	}

	// Apply pagination
	offset := (params.PageNumber - 1) * params.PerPageNumber
	query = query.Offset(offset).Limit(params.PerPageNumber)

	// Execute query with preloads
	result := query.Preload("Molecules", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, canonical_smiles, identifier") // Only select necessary fields to prevent loops
	}).Find(&locations)

	if result.Error != nil {
		ErrorResponse(c, 500, "Failed to fetch locations")
		return
	}

	// Marshal the response using our custom marshaler
	jsonData, err := models.MarshalToJSON(gin.H{
		"locations": locations,
		"total":    total,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

// GetLocationByID handles GET /api/v1/locations/:id
func GetLocationByID(c *gin.Context) {
	id := c.Param("id")
	var location models.GeoLocation

	result := db.Preload("Molecules", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, canonical_smiles, identifier") // Only select necessary fields to prevent loops
	}).First(&location, id)

	if result.Error != nil {
		ErrorResponse(c, 404, "Location not found")
		return
	}

	SuccessResponse(c, location)
}

// GetMoleculesByLocation handles GET /api/v1/locations/:id/molecules
func GetMoleculesByLocation(c *gin.Context) {
	id := c.Param("id")
	params := ParseQueryParams(c)
	var location models.GeoLocation
	var total int64

	// First check if location exists
	if err := db.First(&location, id).Error; err != nil {
		ErrorResponse(c, 404, "Location not found")
		return
	}

	// Build query for molecules
	query := db.Model(&models.Molecule{}).
		Joins("JOIN molecule_geo_location ON molecule_geo_location.molecule_id = molecules.id").
		Where("molecule_geo_location.geo_location_id = ?", id)

	// Apply search if provided
	if params.Search != "" {
		query = query.Where("molecules.name ILIKE ? OR molecules.canonical_smiles ILIKE ? OR molecules.identifier ILIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	// Get total count
	query.Count(&total)

	// Apply ordering
	if params.OrderByString != "" {
		order := params.OrderByString
		if params.OrderDir == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	}

	// Apply pagination
	offset := (params.PageNumber - 1) * params.PerPageNumber
	query = query.Offset(offset).Limit(params.PerPageNumber)

	// Execute query with preloads
	var molecules []models.Molecule
	result := query.Preload("Properties").
		Preload("Organisms").
		Preload("GeoLocations").
		Find(&molecules)

	if result.Error != nil {
		ErrorResponse(c, 500, "Failed to fetch molecules for location")
		return
	}

	PaginatedSuccessResponse(c, molecules, total, params.PageNumber)
}

// GetOBISLocations handles GET /api/v1/obis/locations
func GetOBISLocations(c *gin.Context) {
	// Get aphia_ids from query parameter
	aphiaIDsStr := c.Query("aphia_ids")
	if aphiaIDsStr == "" {
		ErrorResponse(c, 400, "Missing required query parameter: aphia_ids")
		return
	}

	// Parse comma-separated aphia_ids
	aphiaIDStrs := strings.Split(aphiaIDsStr, ",")
	var aphiaIDs []int
	for _, idStr := range aphiaIDStrs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			ErrorResponse(c, 400, "Invalid aphia_id format")
			return
		}
		aphiaIDs = append(aphiaIDs, id)
	}

	// Validate that all aphiaids exist in organisms table
	var count int64
	for _, aphiaID := range aphiaIDs {
		result := db.Model(&models.Organism{}).Where("aphiaid_worms = ?", aphiaID).Count(&count)
		if result.Error != nil {
			log.Printf("Database error while validating aphiaid %d: %v", aphiaID, result.Error)
			ErrorResponse(c, 500, fmt.Sprintf("Database error while validating aphiaid %d: %v", aphiaID, result.Error))
			return
		}
		if count == 0 {
			ErrorResponse(c, 400, fmt.Sprintf("AphiaID %d not found in organisms table", aphiaID))
			return
		}
	}

	// Process all aphiaids and combine results
	var allResults []map[string]interface{}
	totalCount := 0

	for _, aphiaID := range aphiaIDs {
		var cache models.OBISCache
		result := db.Where("aphiaid_worms = ?", aphiaID).First(&cache)
		
		var obisData map[string]interface{}
		
		if result.Error == nil {
			// Found in cache
			if err := json.Unmarshal([]byte(cache.OBISData), &obisData); err != nil {
				log.Printf("Failed to parse cached data for aphiaid %d: %v", aphiaID, err)
				ErrorResponse(c, 500, fmt.Sprintf("Failed to parse cached data for aphiaid %d: %v", aphiaID, err))
				return
			}
		} else if result.Error == gorm.ErrRecordNotFound {
			// Not found in cache, fetch from OBIS API
			url := fmt.Sprintf("https://api.obis.org/v3/occurrence?taxonid=%d&fields=aphiaID,date_mid,decimalLatitude,decimalLongitude,id&size=10000", aphiaID)
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Failed to fetch from OBIS API for aphiaid %d: %v", aphiaID, err)
				ErrorResponse(c, 500, fmt.Sprintf("Failed to fetch from OBIS API for aphiaid %d: %v", aphiaID, err))
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read OBIS API response for aphiaid %d: %v", aphiaID, err)
				ErrorResponse(c, 500, fmt.Sprintf("Failed to read OBIS API response for aphiaid %d: %v", aphiaID, err))
				return
			}

			// Cache the result
			cache := models.OBISCache{
				AphiaIDWorms: aphiaID,
				OBISData:    string(body),
				CreatedAt:   models.SQLiteTime(time.Now()),
				UpdatedAt:   models.SQLiteTime(time.Now()),
			}

			if err := db.Create(&cache).Error; err != nil {
				log.Printf("Failed to cache OBIS data for aphiaid %d: %v", aphiaID, err)
				ErrorResponse(c, 500, fmt.Sprintf("Failed to cache OBIS data for aphiaid %d: %v", aphiaID, err))
				return
			}

			if err := json.Unmarshal(body, &obisData); err != nil {
				log.Printf("Failed to parse OBIS API response for aphiaid %d: %v", aphiaID, err)
				ErrorResponse(c, 500, fmt.Sprintf("Failed to parse OBIS API response for aphiaid %d: %v", aphiaID, err))
				return
			}
		} else {
			log.Printf("Database error while checking cache for aphiaid %d: %v", aphiaID, result.Error)
			ErrorResponse(c, 500, fmt.Sprintf("Database error while checking cache for aphiaid %d: %v", aphiaID, result.Error))
			return
		}

		// Transform and append results for this aphiaid
		if results, ok := obisData["results"].([]interface{}); ok {
			totalCount += len(results)
			for _, r := range results {
				if record, ok := r.(map[string]interface{}); ok {
					// Convert milliseconds to time.Time and format as YYYY-MM-DD
					var formattedDate string
					if dateMid, ok := record["date_mid"].(float64); ok {
						t := time.Unix(0, int64(dateMid)*int64(time.Millisecond))
						formattedDate = t.Format("2006-01-02")
					}

					transformedRecord := map[string]interface{}{
						"name": record["id"],
						"value": []interface{}{
							record["decimalLongitude"].(float64),
							record["decimalLatitude"].(float64),
							formattedDate,
						},
					}
					allResults = append(allResults, transformedRecord)
				}
			}
		}
	}

	// Return combined results
	SuccessResponse(c, map[string]interface{}{
		"total":   totalCount,
		"results": allResults,
	})
}

// transformOBISData transforms the OBIS API response into a map-friendly format
func transformOBISData(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Extract total count
	if results, ok := data["results"].([]interface{}); ok {
		result["total"] = len(results)
		
		// Transform results into the requested format
		transformedResults := make([]map[string]interface{}, 0, len(results))
		for _, r := range results {
			if record, ok := r.(map[string]interface{}); ok {
				// Convert milliseconds to time.Time and format as YYYY-MM-DD
				var formattedDate string
				if dateMid, ok := record["date_mid"].(float64); ok {
					t := time.Unix(0, int64(dateMid)*int64(time.Millisecond))
					formattedDate = t.Format("2006-01-02")
				}

				transformedRecord := map[string]interface{}{
					"name": record["id"],
					"value": []interface{}{
						record["decimalLongitude"].(float64),
						record["decimalLatitude"].(float64),
						formattedDate,
					},
				}
				transformedResults = append(transformedResults, transformedRecord)
			}
		}
		result["results"] = transformedResults
	}

	return result
} 