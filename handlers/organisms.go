/*
 * MarineNP Organisms Handlers
 * Purpose: HTTP handlers for managing and retrieving organism data
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file provides endpoints for accessing and managing biological organisms
 * associated with marine natural products in the MarineNP database.
 */

package handlers

import (
	"marinenp/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetOrganisms handles GET /api/v1/organisms
func GetOrganisms(c *gin.Context) {
	params := ParseQueryParams(c)
	var organisms []models.Organism
	var total int64

	// Build query
	query := db.Model(&models.Organism{}).Where("is_marine = TRUE")

	// Apply search if provided
	if params.Search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(params.Search)+"%")
	}

	// Get total count
	query.Count(&total)

	// Apply ordering
	if params.OrderByString != "" {
		order := params.OrderByString
		if params.OrderDir == "desc" {
			order += " DESC"
		}
		query = query.Order("organisms.id ASC").Order(order)
	} else {
		query = query.Order("organisms.id ASC")
	}

	// Apply pagination
	offset := (params.PageNumber - 1) * params.PerPageNumber
	query = query.Offset(offset).Limit(params.PerPageNumber)

	// Execute query with preloads
	result := query.Preload("Molecules", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, canonical_smiles, identifier") // Only select necessary fields to prevent loops
	}).Find(&organisms)

	if result.Error != nil {
		ErrorResponse(c, 500, "Failed to fetch organisms")
		return
	}

	// Marshal the response using our custom marshaler
	jsonData, err := models.MarshalToJSON(gin.H{
		"organisms": organisms,
		"total":    total,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

// GetOrganismByID handles GET /api/v1/organisms/:id
func GetOrganismByID(c *gin.Context) {
	id := c.Param("id")
	var organism models.Organism

	result := db.Preload("Molecules", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, canonical_smiles, identifier") // Only select necessary fields to prevent loops
	}).First(&organism, id)

	if result.Error != nil {
		ErrorResponse(c, 404, "Organism not found")
		return
	}

	SuccessResponse(c, organism)
}

// GetMoleculesByOrganism handles GET /api/v1/organisms/:id/molecules
func GetMoleculesByOrganism(c *gin.Context) {
	id := c.Param("id")
	params := ParseQueryParams(c)
	var organism models.Organism
	var total int64

	// First check if organism exists
	if err := db.First(&organism, id).Error; err != nil {
		ErrorResponse(c, 404, "Organism not found")
		return
	}

	// Build query for molecules
	query := db.Model(&models.Molecule{}).
		Joins("JOIN molecule_organism ON molecule_organism.molecule_id = molecules.id").
		Where("molecule_organism.organism_id = ? AND molecules.is_marine = TRUE", id)

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
		ErrorResponse(c, 500, "Failed to fetch molecules for organism")
		return
	}

	PaginatedSuccessResponse(c, molecules, total, params.PageNumber)
}

// GetOrganismsAutocomplete handles GET /api/v1/organisms/autocomplete
func GetOrganismsAutocomplete(c *gin.Context) {
	// Format response for autocomplete
	type AutocompleteOption struct {
		Label string `json:"label"`
		Value int64  `json:"value"`
	}

	search := c.Query("search")
	if search == "" {
		SuccessResponse(c, []AutocompleteOption{})
		return
	}

	var organisms []models.Organism
	query := db.Model(&models.Organism{}).
		Where("is_marine = TRUE").
		Where("LOWER(name) LIKE ? OR LOWER(name_aphia_worms) LIKE ?", 
			"%"+strings.ToLower(search)+"%", 
			"%"+strings.ToLower(search)+"%")

	// Limit results to 10 for autocomplete
	result := query.Limit(10).Find(&organisms)

	if result.Error != nil {
		ErrorResponse(c, 500, "Failed to fetch organisms")
		return
	}

	var options []AutocompleteOption
	for _, org := range organisms {
		options = append(options, AutocompleteOption{
			Label: org.Name,
			Value: int64(*org.AphiaIDWorms),
		})
	}

	SuccessResponse(c, options)
} 