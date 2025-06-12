/*
 * MarineNP Collections Handlers
 * Purpose: HTTP handlers for managing and retrieving collection data
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file provides endpoints for accessing and managing curated collections
 * of marine natural products in the MarineNP database.
 */

package handlers

import (
	"marinenp/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetCollections handles GET /api/v1/collections
func GetCollections(c *gin.Context) {
	params := ParseQueryParams(c)
	var collections []models.Collection
	var total int64

	// Build query
	query := db.Model(&models.Collection{})

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
		query = query.Order("collections.id ASC").Order(order)
	} else {
		query = query.Order("collections.id ASC")
	}

	// Apply pagination
	offset := (params.PageNumber - 1) * params.PerPageNumber
	query = query.Offset(offset).Limit(params.PerPageNumber)

	// Execute query without preloading molecules
	result := query.Find(&collections)

	if result.Error != nil {
		ErrorResponse(c, 500, "Failed to fetch collections")
		return
	}

	// Marshal the response using our custom marshaler
	jsonData, err := models.MarshalToJSON(gin.H{
		"collections": collections,
		"total":      total,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

// GetCollectionByID handles GET /api/v1/collections/:id
func GetCollectionByID(c *gin.Context) {
	id := c.Param("id")
	var collection models.Collection

	// Get collection without preloading molecules
	result := db.First(&collection, id)

	if result.Error != nil {
		ErrorResponse(c, 404, "Collection not found")
		return
	}

	SuccessResponse(c, collection)
}

// GetMoleculesByCollection handles GET /api/v1/collections/:id/molecules
func GetMoleculesByCollection(c *gin.Context) {
	id := c.Param("id")
	params := ParseQueryParams(c)
	var collection models.Collection
	var total int64

	// First check if collection exists
	if err := db.First(&collection, id).Error; err != nil {
		ErrorResponse(c, 404, "Collection not found")
		return
	}

	// Build query for molecules
	query := db.Model(&models.Molecule{}).
		Joins("JOIN collection_molecule ON collection_molecule.molecule_id = molecules.id").
		Where("collection_molecule.collection_id = ?", id)

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
		ErrorResponse(c, 500, "Failed to fetch molecules for collection")
		return
	}

	PaginatedSuccessResponse(c, molecules, total, params.PageNumber)
} 