/*
 * MarineNP Molecules Handlers
 * Purpose: HTTP handlers for managing and retrieving molecule data
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file provides endpoints for accessing and managing molecular data
 * in the MarineNP database, including search, export, and analysis features.
 */

package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"marinenp/models"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetMoleculeByID handles GET /api/v1/molecules/:identifier
func GetMoleculeByID(c *gin.Context) {
	identifier := c.Param("identifier")
	var molecule models.Molecule

	result := db.Preload("Properties").
		Preload("Organisms", "is_marine = TRUE").
		Preload("GeoLocations").
		Where("identifier = ?", identifier).
		First(&molecule)

	if result.Error != nil {
		ErrorResponse(c, 404, "Molecule not found")
		return
	}

	SuccessResponse(c, molecule)
}

// GetPropertyRanges handles GET /api/v1/molecules/properties/ranges
func GetPropertyRanges(c *gin.Context) {
	var ranges struct {
		TotalAtomCount              struct{ Min, Max int }     `json:"total_atom_count"`
		HeavyAtomCount             struct{ Min, Max int }     `json:"heavy_atom_count"`
		MolecularWeight            struct{ Min, Max float64 } `json:"molecular_weight"`
		ExactMolecularWeight       struct{ Min, Max float64 } `json:"exact_molecular_weight"`
		Alogp                      struct{ Min, Max float64 } `json:"alogp"`
		TopologicalPolarSurfaceArea struct{ Min, Max float64 } `json:"topological_polar_surface_area"`
		RotatableBondCount         struct{ Min, Max int }     `json:"rotatable_bond_count"`
		HydrogenBondAcceptors      struct{ Min, Max int }     `json:"hydrogen_bond_acceptors"`
		HydrogenBondDonors         struct{ Min, Max int }     `json:"hydrogen_bond_donors"`
	}

	// Get ranges for each property
	db.Model(&models.Properties{}).Select("MIN(total_atom_count) as min, MAX(total_atom_count) as max").Scan(&ranges.TotalAtomCount)
	db.Model(&models.Properties{}).Select("MIN(heavy_atom_count) as min, MAX(heavy_atom_count) as max").Scan(&ranges.HeavyAtomCount)
	db.Model(&models.Properties{}).Select("MIN(molecular_weight) as min, MAX(molecular_weight) as max").Scan(&ranges.MolecularWeight)
	db.Model(&models.Properties{}).Select("MIN(exact_molecular_weight) as min, MAX(exact_molecular_weight) as max").Scan(&ranges.ExactMolecularWeight)
	db.Model(&models.Properties{}).Select("MIN(alogp) as min, MAX(alogp) as max").Scan(&ranges.Alogp)
	db.Model(&models.Properties{}).Select("MIN(topological_polar_surface_area) as min, MAX(topological_polar_surface_area) as max").Scan(&ranges.TopologicalPolarSurfaceArea)
	db.Model(&models.Properties{}).Select("MIN(rotatable_bond_count) as min, MAX(rotatable_bond_count) as max").Scan(&ranges.RotatableBondCount)
	db.Model(&models.Properties{}).Select("MIN(hydrogen_bond_acceptors) as min, MAX(hydrogen_bond_acceptors) as max").Scan(&ranges.HydrogenBondAcceptors)
	db.Model(&models.Properties{}).Select("MIN(hydrogen_bond_donors) as min, MAX(hydrogen_bond_donors) as max").Scan(&ranges.HydrogenBondDonors)

	SuccessResponse(c, ranges)
}

// SearchMolecules handles GET /api/v1/molecules/search
func SearchMolecules(c *gin.Context) {
	params := ParseQueryParams(c)
	var molecules []models.Molecule
	var total int64

	// Build the base query, keeping the original constraint.
	query := db.Model(&models.Molecule{})

	// --- Dynamic Query Parameter Parsing ---
	// Get all query parameters from the URL to handle flexible filtering.
	queryParams := c.Request.URL.Query()

	// Debug: Print all query parameters
	fmt.Printf("Query Parameters: %+v\n", queryParams)

	// Handle conditions array if present
	conditions := make([]struct {
		Field    string `json:"field"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	}, 0)

	// Parse conditions from query parameters
	i := 0
	for {
		field := queryParams.Get(fmt.Sprintf("conditions[%d][field]", i))
		operator := queryParams.Get(fmt.Sprintf("conditions[%d][operator]", i))
		value := queryParams.Get(fmt.Sprintf("conditions[%d][value]", i))

		if field == "" && operator == "" && value == "" {
			break
		}

		conditions = append(conditions, struct {
			Field    string `json:"field"`
			Operator string `json:"operator"`
			Value    string `json:"value"`
		}{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
		i++
	}

	fmt.Printf("Parsed conditions: %+v\n", conditions)

	// Check if we need to join with properties table
	needsPropertiesJoin := false
	needsOrganismJoin := false
	for _, condition := range conditions {
		if strings.HasPrefix(condition.Field, "properties.") {
			needsPropertiesJoin = true
		}
		if condition.Field == "organism" || condition.Field == "organism_id" {
			needsOrganismJoin = true
		}
	}

	// Join with properties table if needed
	if needsPropertiesJoin {
		query = query.Joins("JOIN properties ON properties.molecule_id = molecules.id")
	}

	// Join with organism table if needed
	if needsOrganismJoin {
		query = query.Joins("JOIN molecule_organism ON molecule_organism.molecule_id = molecules.id").
			Joins("JOIN organisms ON organisms.id = molecule_organism.organism_id")
		query = query.Where("organisms.is_marine = TRUE")
	} else {
		query = query.Where("molecules.is_marine = TRUE")
	}

	// Apply each condition to the query
	for _, condition := range conditions {
		// Skip if any part is missing
		if condition.Field == "" || condition.Operator == "" || condition.Value == "" {
			fmt.Printf("Skipping condition due to missing parts: %+v\n", condition)
			continue
		}

		// Handle organism filter
		if condition.Field == "organism" {
			sqlOperator := ""
			switch condition.Operator {
			case "eq":
				sqlOperator = "="
			case "ne":
				sqlOperator = "!="
			case "contains":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value) + "%"
			case "startsWith":
				sqlOperator = "LIKE"
				condition.Value = strings.ToLower(condition.Value) + "%"
			case "endsWith":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value)
			default:
				fmt.Printf("Invalid operator for organism filter: %s\n", condition.Operator)
				continue
			}
			query = query.Where(
				"(LOWER(organisms.name) "+sqlOperator+" ? OR "+
				"LOWER(organisms.iri) "+sqlOperator+" ? OR "+
				"LOWER(organisms.slug) "+sqlOperator+" ? OR "+
				"LOWER(organisms.name_aphia_worms) "+sqlOperator+" ?)",
				condition.Value, condition.Value, condition.Value, condition.Value)
			continue
		}

		// Handle organism_id filter
		if condition.Field == "organism_id" {
			sqlOperator := ""
			switch condition.Operator {
			case "eq":
				sqlOperator = "="
			case "ne":
				sqlOperator = "!="
			default:
				fmt.Printf("Invalid operator for organism_id filter: %s\n", condition.Operator)
				continue
			}
			// Convert value to integer
			organismID, err := strconv.Atoi(condition.Value)
			if err != nil {
				fmt.Printf("Invalid organism_id value: %s\n", condition.Value)
				continue
			}
			// Use the integer value directly
			query = query.Where("molecule_organism.organism_id "+sqlOperator+" ?", organismID)
			continue
		}

		// Handle property filters
		if strings.HasPrefix(condition.Field, "properties.") {
			propertyField := strings.TrimPrefix(condition.Field, "properties.")
			
			sqlOperator := ""
			switch condition.Operator {
			case "eq":
				sqlOperator = "="
			case "ne":
				sqlOperator = "!="
			case "lt":
				sqlOperator = "<"
			case "lte":
				sqlOperator = "<="
			case "gt":
				sqlOperator = ">"
			case "gte":
				sqlOperator = ">="
			case "contains":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value) + "%"
			case "startsWith":
				sqlOperator = "LIKE"
				condition.Value = strings.ToLower(condition.Value) + "%"
			case "endsWith":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value)
			default:
				fmt.Printf("Invalid operator for property filter: %s\n", condition.Operator)
				continue
			}

			// Check if the field is numeric
			isNumeric := false
			switch propertyField {
			case "heavy_atom_count", "total_atom_count", "rotatable_bond_count",
				"hydrogen_bond_acceptors", "hydrogen_bond_donors",
				"hydrogen_bond_acceptors_lipinski", "hydrogen_bond_donors_lipinski",
				"lipinski_rule_of_five_violations", "aromatic_rings_count",
				"number_of_minimal_rings", "molecular_weight", "exact_molecular_weight",
				"alogp", "topological_polar_surface_area", "formal_charge",
				"van_der_walls_volume", "qed_drug_likeliness", "np_likeness",
				"fractioncsp3":
				isNumeric = true
			}

			if isNumeric {
				// For numeric fields, use CAST to ensure proper numeric comparison
				query = query.Where("CAST(properties."+propertyField+" AS REAL) "+sqlOperator+" ?", condition.Value)
			} else {
				query = query.Where("properties."+propertyField+" "+sqlOperator+" ?", condition.Value)
			}
		} else {
			// Handle direct field filters
			sqlOperator := ""
			switch condition.Operator {
			case "eq":
				sqlOperator = "="
			case "ne":
				sqlOperator = "!="
			case "lt":
				sqlOperator = "<"
			case "lte":
				sqlOperator = "<="
			case "gt":
				sqlOperator = ">"
			case "gte":
				sqlOperator = ">="
			case "contains":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value) + "%"
			case "startsWith":
				sqlOperator = "LIKE"
				condition.Value = strings.ToLower(condition.Value) + "%"
			case "endsWith":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value)
			default:
				fmt.Printf("Invalid operator for direct field filter: %s\n", condition.Operator)
				continue
			}

			// Check if the field is numeric
			isNumeric := false
			switch condition.Field {
			case "id", "organism_count", "geo_count", "citation_count", "collection_count",
				"synonym_count", "variants_count":
				isNumeric = true
			}

			if isNumeric {
				// For numeric fields, use CAST to ensure proper numeric comparison
				query = query.Where("CAST("+condition.Field+" AS INTEGER) "+sqlOperator+" ?", condition.Value)
			} else {
				query = query.Where("LOWER("+condition.Field+") "+sqlOperator+" ?", condition.Value)
			}
		}
	}

	// Handle keyword search if present
	if keyword := queryParams.Get("keyword"); keyword != "" {
		searchValue := "%" + strings.ToLower(keyword) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR " +
			"LOWER(canonical_smiles) LIKE ? OR " +
			"LOWER(identifier) LIKE ? OR " +
			"LOWER(cas) LIKE ? OR " +
			"LOWER(synonyms) LIKE ? OR " +
			"LOWER(iupac_name) LIKE ? OR " +
			"LOWER(standard_inchi) LIKE ? OR " +
			"LOWER(standard_inchi_key) LIKE ?",
			searchValue, searchValue, searchValue, searchValue, searchValue,
			searchValue, searchValue, searchValue)
	}

	// Debug: Print the final SQL query
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&models.Molecule{})
	})
	fmt.Printf("Final SQL Query: %s\n", sql)

	// Get total count *after* all filtering has been applied
	if err := query.Count(&total).Error; err != nil {
		ErrorResponse(c, 500, "Failed to count molecules")
		return
	}

	// Apply ordering using parameters from the helper function
	if params.OrderByString != "" {
		order := params.OrderByString
		if params.OrderDir == "desc" {
			order += " DESC"
		}
		query = query.Order("molecules.id ASC").Order(order)
	} else {
		query = query.Order("molecules.id ASC")
	}

	// Apply pagination using parameters from the helper function
	offset := (params.PageNumber - 1) * params.PerPageNumber
	query = query.Offset(offset).Limit(params.PerPageNumber)

	// Execute the final query with preloads
	result := query.Preload("Properties").
		Preload("Organisms").
		Preload("GeoLocations").
		Find(&molecules)

	if result.Error != nil {
		ErrorResponse(c, 500, "Failed to search molecules")
		return
	}

	// Marshal the response using our custom marshaler
	jsonData, err := models.MarshalToJSON(gin.H{
		"molecules": molecules,
		"total":    total,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal response"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

// ExportMolecules handles GET /api/v1/molecules/export
func ExportMolecules(c *gin.Context) {
	params := ParseQueryParams(c)

	// Build the base query, keeping the original constraint.
	query := db.Model(&models.Molecule{})

	// --- Dynamic Query Parameter Parsing ---
	// Get all query parameters from the URL to handle flexible filtering.
	queryParams := c.Request.URL.Query()

	// Handle conditions array if present
	conditions := make([]struct {
		Field    string `json:"field"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	}, 0)

	// Parse conditions from query parameters
	i := 0
	for {
		field := queryParams.Get(fmt.Sprintf("conditions[%d][field]", i))
		operator := queryParams.Get(fmt.Sprintf("conditions[%d][operator]", i))
		value := queryParams.Get(fmt.Sprintf("conditions[%d][value]", i))

		if field == "" && operator == "" && value == "" {
			break
		}

		conditions = append(conditions, struct {
			Field    string `json:"field"`
			Operator string `json:"operator"`
			Value    string `json:"value"`
		}{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
		i++
	}

	// Check if we need to join with properties table
	needsPropertiesJoin := false
	needsOrganismJoin := false
	for _, condition := range conditions {
		if strings.HasPrefix(condition.Field, "properties.") {
			needsPropertiesJoin = true
		}
		if condition.Field == "organism" || condition.Field == "organism_id" {
			needsOrganismJoin = true
		}
	}

	// Join with properties table if needed
	if needsPropertiesJoin {
		query = query.Joins("JOIN properties ON properties.molecule_id = molecules.id")
	}

	// Join with organism table if needed
	if needsOrganismJoin {
		query = query.Joins("JOIN molecule_organism ON molecule_organism.molecule_id = molecules.id").
			Joins("JOIN organisms ON organisms.id = molecule_organism.organism_id")
		query = query.Where("organisms.is_marine = TRUE")
	} else {
		query = query.Where("molecules.is_marine = TRUE")
	}

	// Process conditions in batches to avoid parameter limit
	const batchSize = 100
	for i := 0; i < len(conditions); i += batchSize {
		end := i + batchSize
		if end > len(conditions) {
			end = len(conditions)
		}

		batch := conditions[i:end]
		for _, condition := range batch {
			if condition.Field == "" || condition.Operator == "" || condition.Value == "" {
				continue
			}

			// Handle organism filter
			if condition.Field == "organism" {
				sqlOperator := ""
				switch condition.Operator {
				case "eq":
					sqlOperator = "="
				case "ne":
					sqlOperator = "!="
				case "contains":
					sqlOperator = "LIKE"
					condition.Value = "%" + strings.ToLower(condition.Value) + "%"
				case "startsWith":
					sqlOperator = "LIKE"
					condition.Value = strings.ToLower(condition.Value) + "%"
				case "endsWith":
					sqlOperator = "LIKE"
					condition.Value = "%" + strings.ToLower(condition.Value)
				default:
					continue
				}
				query = query.Where(
					"(LOWER(organisms.name) "+sqlOperator+" ? OR "+
					"LOWER(organisms.iri) "+sqlOperator+" ? OR "+
					"LOWER(organisms.slug) "+sqlOperator+" ? OR "+
					"LOWER(organisms.name_aphia_worms) "+sqlOperator+" ?)",
					condition.Value, condition.Value, condition.Value, condition.Value)
				continue
			}

			// Handle organism_id filter
			if condition.Field == "organism_id" {
				sqlOperator := ""
				switch condition.Operator {
				case "eq":
					sqlOperator = "="
				case "ne":
					sqlOperator = "!="
				default:
					fmt.Printf("Invalid operator for organism_id filter: %s\n", condition.Operator)
					continue
				}
				// Convert value to integer
				organismID, err := strconv.Atoi(condition.Value)
				if err != nil {
					fmt.Printf("Invalid organism_id value: %s\n", condition.Value)
					continue
				}
				// Use the integer value directly
				query = query.Where("molecule_organism.organism_id "+sqlOperator+" ?", organismID)
				continue
			}

			// Handle property filters
			if strings.HasPrefix(condition.Field, "properties.") {
				propertyField := strings.TrimPrefix(condition.Field, "properties.")
				
				sqlOperator := ""
				switch condition.Operator {
				case "eq":
					sqlOperator = "="
				case "ne":
					sqlOperator = "!="
				case "lt":
					sqlOperator = "<"
				case "lte":
					sqlOperator = "<="
				case "gt":
					sqlOperator = ">"
				case "gte":
					sqlOperator = ">="
				case "contains":
					sqlOperator = "LIKE"
					condition.Value = "%" + strings.ToLower(condition.Value) + "%"
				case "startsWith":
					sqlOperator = "LIKE"
					condition.Value = strings.ToLower(condition.Value) + "%"
				case "endsWith":
					sqlOperator = "LIKE"
					condition.Value = "%" + strings.ToLower(condition.Value)
				default:
					fmt.Printf("Invalid operator for property filter: %s\n", condition.Operator)
					continue
				}

				// Check if the field is numeric
				isNumeric := false
				switch propertyField {
				case "heavy_atom_count", "total_atom_count", "rotatable_bond_count",
					"hydrogen_bond_acceptors", "hydrogen_bond_donors",
					"hydrogen_bond_acceptors_lipinski", "hydrogen_bond_donors_lipinski",
					"lipinski_rule_of_five_violations", "aromatic_rings_count",
					"number_of_minimal_rings", "molecular_weight", "exact_molecular_weight",
					"alogp", "topological_polar_surface_area", "formal_charge",
					"van_der_walls_volume", "qed_drug_likeliness", "np_likeness",
					"fractioncsp3":
					isNumeric = true
				}

				if isNumeric {
					// For numeric fields, use CAST to ensure proper numeric comparison
					query = query.Where("CAST(properties."+propertyField+" AS REAL) "+sqlOperator+" ?", condition.Value)
				} else {
					query = query.Where("properties."+propertyField+" "+sqlOperator+" ?", condition.Value)
				}
			} else {
				// Handle direct field filters
				sqlOperator := ""
				switch condition.Operator {
				case "eq":
					sqlOperator = "="
				case "ne":
					sqlOperator = "!="
				case "lt":
					sqlOperator = "<"
				case "lte":
					sqlOperator = "<="
				case "gt":
					sqlOperator = ">"
				case "gte":
					sqlOperator = ">="
				case "contains":
					sqlOperator = "LIKE"
					condition.Value = "%" + strings.ToLower(condition.Value) + "%"
				case "startsWith":
					sqlOperator = "LIKE"
					condition.Value = strings.ToLower(condition.Value) + "%"
				case "endsWith":
					sqlOperator = "LIKE"
					condition.Value = "%" + strings.ToLower(condition.Value)
				default:
					fmt.Printf("Invalid operator for direct field filter: %s\n", condition.Operator)
					continue
				}

				// Check if the field is numeric
				isNumeric := false
				switch condition.Field {
				case "id", "organism_count", "geo_count", "citation_count", "collection_count",
					"synonym_count", "variants_count":
					isNumeric = true
				}

				if isNumeric {
					// For numeric fields, use CAST to ensure proper numeric comparison
					query = query.Where("CAST("+condition.Field+" AS INTEGER) "+sqlOperator+" ?", condition.Value)
				} else {
					query = query.Where("LOWER("+condition.Field+") "+sqlOperator+" ?", condition.Value)
				}
			}
		}
	}

	// Handle keyword search if present
	if keyword := queryParams.Get("keyword"); keyword != "" {
		searchValue := "%" + strings.ToLower(keyword) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR " +
			"LOWER(canonical_smiles) LIKE ? OR " +
			"LOWER(identifier) LIKE ? OR " +
			"LOWER(cas) LIKE ? OR " +
			"LOWER(synonyms) LIKE ? OR " +
			"LOWER(iupac_name) LIKE ? OR " +
			"LOWER(standard_inchi) LIKE ? OR " +
			"LOWER(standard_inchi_key) LIKE ?",
			searchValue, searchValue, searchValue, searchValue, searchValue,
			searchValue, searchValue, searchValue)
	}

	// Apply ordering
	if params.OrderByString != "" {
		order := params.OrderByString
		if params.OrderDir == "desc" {
			order += " DESC"
		}
		query = query.Order("molecules.id ASC").Order(order)
	} else {
		query = query.Order("molecules.id ASC")
	}

	// Create a buffer for CSV data
	var csvBuffer bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuffer)

	// Define organized headers in logical groups
	headers := []string{
		// Basic Information
		"identifier",
		"name",
		"cas",
		"iupac_name",
		"synonyms",
		"molecular_formula",
		
		// Structure Information
		"canonical_smiles",
		"sugar_free_smiles",
		"standard_inchi",
		"standard_inchi_key",
		"murcko_framework",
		"has_stereo",
		
		// Physical Properties
		"molecular_weight",
		"exact_molecular_weight",
		"total_atom_count",
		"heavy_atom_count",
		"rotatable_bond_count",
		"hydrogen_bond_acceptors",
		"hydrogen_bond_donors",
		"topological_polar_surface_area",
		"alogp",
		"formal_charge",
		"van_der_walls_volume",
		
		// Chemical Classification
		"chemical_class",
		"chemical_sub_class",
		"chemical_super_class",
		"direct_parent_classification",
		"np_classifier_class",
		"np_classifier_superclass",
		"np_classifier_pathway",
		
		// Drug-like Properties
		"lipinski_rule_of_five_violations",
		"hydrogen_bond_acceptors_lipinski",
		"hydrogen_bond_donors_lipinski",
		"qed_drug_likeliness",
		"np_likeness",
		
		// Sugar Information
		"contains_sugar",
		"contains_linear_sugars",
		"contains_ring_sugars",
		"np_classifier_is_glycoside",
		
		// Ring Information
		"aromatic_rings_count",
		"number_of_minimal_rings",
		"fractioncsp3",
		
		// Status and Metadata
		"status",
		"is_marine",
		"citation_count",
		"organism_count",
		"geo_count",
	}

	// Write header
	if err := csvWriter.Write(headers); err != nil {
		ErrorResponse(c, 500, "Failed to write CSV header")
		return
	}

	// Process in chunks
	const chunkSize = 50000
	offset := 0

	// Create a fresh query for export using the model
	exportQuery := db.Model(&models.Molecule{}).
		Select("molecules.*, properties.*").
		Joins("LEFT JOIN properties ON properties.molecule_id = molecules.id")

	// Check if we need organism join
	if needsOrganismJoin {
		exportQuery = exportQuery.
			Joins("LEFT JOIN molecule_organism ON molecule_organism.molecule_id = molecules.id").
			Joins("LEFT JOIN organisms ON organisms.id = molecule_organism.organism_id")
	}
	
	// Apply the same conditions from the original query
	exportQuery = exportQuery.Where(query.Statement.Clauses["WHERE"].Expression)

	for {
		var records []map[string]interface{}
		
		// Get a chunk of records using the export query
		result := exportQuery.Offset(offset).Limit(chunkSize).Find(&records)

		if result.Error != nil {
			ErrorResponse(c, 500, "Failed to export molecules")
			return
		}

		// If no more results, break
		if len(records) == 0 {
			break
		}

		// Write data rows for this chunk
		for _, record := range records {
			row := make([]string, len(headers))
			for i, header := range headers {
				value := record[header]
				if value != nil {
					row[i] = fmt.Sprintf("%v", value)
				}
			}
			if err := csvWriter.Write(row); err != nil {
				ErrorResponse(c, 500, "Failed to write CSV row")
				return
			}
		}

		// Flush after each chunk
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			ErrorResponse(c, 500, "Failed to flush CSV data")
			return
		}
		
		// Move to next chunk
		offset += chunkSize
	}

	// Final flush of CSV data
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		ErrorResponse(c, 500, "Failed to finalize CSV data")
		return
	}

	// Set headers for zip download
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment;filename=molecules_%s.zip", time.Now().Format("20060102_150405")))

	// Create zip writer
	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	// Create search-query.txt in the zip
	queryFile, err := zipWriter.Create("search-query.txt")
	if err != nil {
		ErrorResponse(c, 500, "Failed to create query file")
		return
	}

	// Write the full search query to the file
	fullQuery := fmt.Sprintf("/api/v1/molecules/search?%s", c.Request.URL.RawQuery)
	if _, err := io.WriteString(queryFile, fullQuery); err != nil {
		ErrorResponse(c, 500, "Failed to write query to file")
		return
	}

	// Create CSV file in the zip
	csvFile, err := zipWriter.Create("molecules.csv")
	if err != nil {
		ErrorResponse(c, 500, "Failed to create zip file")
		return
	}

	// Write the CSV buffer to the zip file
	if _, err := io.Copy(csvFile, &csvBuffer); err != nil {
		ErrorResponse(c, 500, "Failed to write CSV to zip")
		return
	}

	// Close the zip writer
	if err := zipWriter.Close(); err != nil {
		ErrorResponse(c, 500, "Failed to close zip file")
		return
	}
}

// AnalyzeMolecules handles GET /api/v1/molecules/analyze
func AnalyzeMolecules(c *gin.Context) {
	parameter := c.Query("parameter")
	chartType := c.Query("chart_type")
	if parameter == "" {
		ErrorResponse(c, 400, "Parameter is required")
		return
	}

	// Build the base query using the same logic as SearchMolecules
	query := db.Model(&models.Molecule{})

	// Get all query parameters from the URL
	queryParams := c.Request.URL.Query()

	// Handle conditions array if present
	conditions := make([]struct {
		Field    string `json:"field"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	}, 0)

	// Parse conditions from query parameters
	i := 0
	for {
		field := queryParams.Get(fmt.Sprintf("conditions[%d][field]", i))
		operator := queryParams.Get(fmt.Sprintf("conditions[%d][operator]", i))
		value := queryParams.Get(fmt.Sprintf("conditions[%d][value]", i))

		if field == "" && operator == "" && value == "" {
			break
		}

		conditions = append(conditions, struct {
			Field    string `json:"field"`
			Operator string `json:"operator"`
			Value    string `json:"value"`
		}{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
		i++
	}

	// Check if we need to join with properties table
	needsPropertiesJoin := false
	needsOrganismJoin := false
	for _, condition := range conditions {
		if strings.HasPrefix(condition.Field, "properties.") {
			needsPropertiesJoin = true
		}
		if condition.Field == "organism" || condition.Field == "organism_id" {
			needsOrganismJoin = true
		}
	}

	// Join with properties table if needed - only once
	if needsPropertiesJoin {
		query = query.Joins("JOIN properties ON properties.molecule_id = molecules.id")
	}

	// Join with organism table if needed
	if needsOrganismJoin {
		query = query.Joins("JOIN molecule_organism ON molecule_organism.molecule_id = molecules.id").
			Joins("JOIN organisms ON organisms.id = molecule_organism.organism_id")
		query = query.Where("organisms.is_marine = TRUE")
	} else {
		query = query.Where("molecules.is_marine = TRUE")
	}

	// Apply each condition to the query
	for _, condition := range conditions {
		// Skip if any part is missing
		if condition.Field == "" || condition.Operator == "" || condition.Value == "" {
			continue
		}

		// Handle organism filter
		if condition.Field == "organism" {
			sqlOperator := ""
			switch condition.Operator {
			case "eq":
				sqlOperator = "="
			case "ne":
				sqlOperator = "!="
			case "contains":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value) + "%"
			case "startsWith":
				sqlOperator = "LIKE"
				condition.Value = strings.ToLower(condition.Value) + "%"
			case "endsWith":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value)
			default:
				continue
			}
			query = query.Where(
				"(LOWER(organisms.name) "+sqlOperator+" ? OR "+
				"LOWER(organisms.iri) "+sqlOperator+" ? OR "+
				"LOWER(organisms.slug) "+sqlOperator+" ? OR "+
				"LOWER(organisms.name_aphia_worms) "+sqlOperator+" ?)",
				condition.Value, condition.Value, condition.Value, condition.Value)
			continue
		}

		// Handle organism_id filter
		if condition.Field == "organism_id" {
			sqlOperator := ""
			switch condition.Operator {
			case "eq":
				sqlOperator = "="
			case "ne":
				sqlOperator = "!="
			default:
				continue
			}
			organismID, err := strconv.Atoi(condition.Value)
			if err != nil {
				continue
			}
			query = query.Where("molecule_organism.organism_id "+sqlOperator+" ?", organismID)
			continue
		}

		// Handle property filters
		if strings.HasPrefix(condition.Field, "properties.") {
			propertyField := strings.TrimPrefix(condition.Field, "properties.")
			
			sqlOperator := ""
			switch condition.Operator {
			case "eq":
				sqlOperator = "="
			case "ne":
				sqlOperator = "!="
			case "lt":
				sqlOperator = "<"
			case "lte":
				sqlOperator = "<="
			case "gt":
				sqlOperator = ">"
			case "gte":
				sqlOperator = ">="
			case "contains":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value) + "%"
			case "startsWith":
				sqlOperator = "LIKE"
				condition.Value = strings.ToLower(condition.Value) + "%"
			case "endsWith":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value)
			default:
				continue
			}

			// Check if the field is numeric
			isNumeric := false
			switch propertyField {
			case "heavy_atom_count", "total_atom_count", "rotatable_bond_count",
				"hydrogen_bond_acceptors", "hydrogen_bond_donors",
				"hydrogen_bond_acceptors_lipinski", "hydrogen_bond_donors_lipinski",
				"lipinski_rule_of_five_violations", "aromatic_rings_count",
				"number_of_minimal_rings", "molecular_weight", "exact_molecular_weight",
				"alogp", "topological_polar_surface_area", "formal_charge",
				"van_der_walls_volume", "qed_drug_likeliness", "np_likeness",
				"fractioncsp3":
				isNumeric = true
			}

			if isNumeric {
				query = query.Where("CAST(properties."+propertyField+" AS REAL) "+sqlOperator+" ?", condition.Value)
			} else {
				query = query.Where("properties."+propertyField+" "+sqlOperator+" ?", condition.Value)
			}
		} else {
			// Handle direct field filters
			sqlOperator := ""
			switch condition.Operator {
			case "eq":
				sqlOperator = "="
			case "ne":
				sqlOperator = "!="
			case "lt":
				sqlOperator = "<"
			case "lte":
				sqlOperator = "<="
			case "gt":
				sqlOperator = ">"
			case "gte":
				sqlOperator = ">="
			case "contains":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value) + "%"
			case "startsWith":
				sqlOperator = "LIKE"
				condition.Value = strings.ToLower(condition.Value) + "%"
			case "endsWith":
				sqlOperator = "LIKE"
				condition.Value = "%" + strings.ToLower(condition.Value)
			default:
				continue
			}

			// Check if the field is numeric
			isNumeric := false
			switch condition.Field {
			case "id", "organism_count", "geo_count", "citation_count", "collection_count",
				"synonym_count", "variants_count":
				isNumeric = true
			}

			if isNumeric {
				query = query.Where("CAST("+condition.Field+" AS INTEGER) "+sqlOperator+" ?", condition.Value)
			} else {
				query = query.Where("LOWER("+condition.Field+") "+sqlOperator+" ?", condition.Value)
			}
		}
	}

	// Handle keyword search if present
	if keyword := queryParams.Get("keyword"); keyword != "" {
		searchValue := "%" + strings.ToLower(keyword) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR " +
			"LOWER(canonical_smiles) LIKE ? OR " +
			"LOWER(identifier) LIKE ? OR " +
			"LOWER(cas) LIKE ? OR " +
			"LOWER(synonyms) LIKE ? OR " +
			"LOWER(iupac_name) LIKE ? OR " +
			"LOWER(standard_inchi) LIKE ? OR " +
			"LOWER(standard_inchi_key) LIKE ?",
			searchValue, searchValue, searchValue, searchValue, searchValue,
			searchValue, searchValue, searchValue)
	}

	// Remove any duplicate joins that might have been added
	query = query.Distinct()

	// Handle different chart types
	switch chartType {
	case "sunburst":
		// First get the filtered molecule IDs
		var moleculeIDs []uint
		if err := query.Select("molecules.id").Find(&moleculeIDs).Error; err != nil {
			ErrorResponse(c, 500, fmt.Sprintf("Failed to get molecule IDs: %v", err))
			return
		}

		// Process molecule IDs in batches to avoid hitting IN clause limits
		const batchSize = 10000
		var result []struct {
			Superclass string `gorm:"column:superclass"`
			Class      string `gorm:"column:class"`
			Pathway    string `gorm:"column:pathway"`
			Count      int64  `gorm:"column:count"`
		}

		// Process in batches
		for i := 0; i < len(moleculeIDs); i += batchSize {
			end := i + batchSize
			if end > len(moleculeIDs) {
				end = len(moleculeIDs)
			}

			batch := moleculeIDs[i:end]
			var batchResult []struct {
				Superclass string `gorm:"column:superclass"`
				Class      string `gorm:"column:class"`
				Pathway    string `gorm:"column:pathway"`
				Count      int64  `gorm:"column:count"`
			}

			if parameter == "classifire" {
				if err := db.Model(&models.Properties{}).
					Select("chemical_super_class as superclass, chemical_class as class, chemical_sub_class as pathway, COUNT(*) as count").
					Where("molecule_id IN ?", batch).
					Group("chemical_super_class, chemical_class, chemical_sub_class").
					Order("count DESC").
					Find(&batchResult).Error; err != nil {
					ErrorResponse(c, 500, fmt.Sprintf("Failed to analyze molecules for batch %d-%d: %v", i, end, err))
					return
				}
			} else if parameter == "np_classifier" {
				if err := db.Model(&models.Properties{}).
					Select("np_classifier_pathway as pathway, np_classifier_superclass as superclass, np_classifier_class as class, COUNT(*) as count").
					Where("molecule_id IN ?", batch).
					Group("np_classifier_pathway, np_classifier_superclass, np_classifier_class").
					Order("count DESC").
					Find(&batchResult).Error; err != nil {
					ErrorResponse(c, 500, fmt.Sprintf("Failed to analyze molecules for batch %d-%d: %v", i, end, err))
					return
				}
			}

			result = append(result, batchResult...)
		}

		// Build sunburst data structure
		sunburstData := make(map[string]map[string]map[string]int64)
		for _, r := range result {
			if parameter == "classifire" {
				if r.Superclass == "" {
					continue
				}
				if _, ok := sunburstData[r.Superclass]; !ok {
					sunburstData[r.Superclass] = make(map[string]map[string]int64)
				}
				if _, ok := sunburstData[r.Superclass][r.Class]; !ok {
					sunburstData[r.Superclass][r.Class] = make(map[string]int64)
				}
				sunburstData[r.Superclass][r.Class][r.Pathway] = r.Count
			} else { // np_classifier
				if r.Pathway == "" {
					continue
				}
				if _, ok := sunburstData[r.Pathway]; !ok {
					sunburstData[r.Pathway] = make(map[string]map[string]int64)
				}
				if _, ok := sunburstData[r.Pathway][r.Superclass]; !ok {
					sunburstData[r.Pathway][r.Superclass] = make(map[string]int64)
				}
				sunburstData[r.Pathway][r.Superclass][r.Class] = r.Count
			}
		}

		// Convert to required format
		var children []map[string]interface{}
		for top, middle := range sunburstData {
			middleChildren := make([]map[string]interface{}, 0)
			for mid, bottom := range middle {
				bottomChildren := make([]map[string]interface{}, 0)
				for bot, count := range bottom {
					if bot != "" {
						bottomChildren = append(bottomChildren, map[string]interface{}{
							"name":  bot,
							"value": count,
						})
					}
				}
				middleData := map[string]interface{}{
					"name": mid,
				}
				if len(bottomChildren) > 0 {
					middleData["children"] = bottomChildren
				} else {
					middleData["value"] = 0
					for _, count := range bottom {
						middleData["value"] = count
						break
					}
				}
				middleChildren = append(middleChildren, middleData)
			}
			children = append(children, map[string]interface{}{
				"name":     top,
				"children": middleChildren,
			})
		}

		SuccessResponse(c, gin.H{
			"values":    children,
			"parameter": parameter,
		})

	case "density":
		// First get the filtered molecule IDs
		var moleculeIDs []uint
		if err := query.Select("molecules.id").Find(&moleculeIDs).Error; err != nil {
			ErrorResponse(c, 500, fmt.Sprintf("Failed to get molecule IDs: %v", err))
			return
		}

		// Process molecule IDs in batches to avoid hitting IN clause limits
		const batchSize = 10000 // Reduced from 65000 to avoid "too many SQL variables" error
		var propertyValues []float64

		for i := 0; i < len(moleculeIDs); i += batchSize {
			end := i + batchSize
			if end > len(moleculeIDs) {
				end = len(moleculeIDs)
			}

			batch := moleculeIDs[i:end]
			var batchValues []float64
			if err := db.Model(&models.Properties{}).
				Where("molecule_id IN ?", batch).
				Pluck("CAST("+parameter+" AS REAL)", &batchValues).Error; err != nil {
				ErrorResponse(c, 500, fmt.Sprintf("Failed to get property values for batch %d-%d: %v", i, end, err))
				return
			}
			propertyValues = append(propertyValues, batchValues...)
		}

		// Find min and max values
		if len(propertyValues) == 0 {
			SuccessResponse(c, gin.H{
				"values":    []struct{}{},
				"parameter": parameter,
			})
			return
		}

		min := propertyValues[0]
		max := propertyValues[0]
		allIntegers := true
		for _, v := range propertyValues {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
			// Check if value is an integer
			if math.Floor(v) != v {
				allIntegers = false
			}
		}

		// Create bins for the histogram
		var numBins int
		var binWidth float64
		var binStarts []float64
		var bins []int64

		if allIntegers {
			// For integer values, use integer bins
			numBins = int(max - min + 1)
			if numBins <= 0 {
				ErrorResponse(c, 500, fmt.Sprintf("Invalid bin count for integer values: min=%v, max=%v", min, max))
				return
			}
			binWidth = 1
			bins = make([]int64, numBins)
			binStarts = make([]float64, numBins)
			for i := 0; i < numBins; i++ {
				binStarts[i] = min + float64(i)
			}
		} else {
			// For floating point values, use 50 evenly spaced bins
			numBins = 50
			binWidth = (max - min) / float64(numBins)
			if binWidth <= 0 {
				ErrorResponse(c, 500, fmt.Sprintf("Invalid bin width for float values: min=%v, max=%v", min, max))
				return
			}
			bins = make([]int64, numBins)
			binStarts = make([]float64, numBins)
			for i := 0; i < numBins; i++ {
				binStarts[i] = min + float64(i)*binWidth
			}
		}

		// Count values in each bin
		for _, value := range propertyValues {
			var binIndex int
			if allIntegers {
				binIndex = int(value - min)
			} else {
				binIndex = int((value - min) / binWidth)
				if binIndex >= numBins {
					binIndex = numBins - 1 // Handle edge case for max value
				}
			}
			if binIndex < 0 || binIndex >= numBins {
				ErrorResponse(c, 500, fmt.Sprintf("Invalid bin index %d for value %v (min=%v, max=%v, numBins=%d)", binIndex, value, min, max, numBins))
				return
			}
			bins[binIndex]++
		}

		// Format the response for the density plot
		values := make([]struct {
			Label string  `json:"label"`
			Value float64 `json:"value"`
		}, numBins)

		for i := 0; i < numBins; i++ {
			// Use the midpoint of the bin as the label
			var midpoint float64
			if allIntegers {
				midpoint = binStarts[i]
			} else {
				midpoint = binStarts[i] + binWidth/2
			}
			values[i].Label = fmt.Sprintf("%.2f", midpoint)
			values[i].Value = float64(bins[i])
		}

		SuccessResponse(c, gin.H{
			"values":    values,
			"parameter": parameter,
		})

	default: // bar
		// For categorical values, count occurrences
		var result []struct {
			Value string `gorm:"column:value"`
			Count int64  `gorm:"column:count"`
		}

		query = query.Select("properties."+parameter+" as value, COUNT(*) as count").
			Group("properties."+parameter).
			Order("count DESC")

		if err := query.Find(&result).Error; err != nil {
			ErrorResponse(c, 500, fmt.Sprintf("Failed to get categorical values: %v", err))
			return
		}

		// Format the response for the bar chart
		values := make([]struct {
			Label string  `json:"label"`
			Value float64 `json:"value"`
		}, len(result))

		for i, r := range result {
			values[i].Label = r.Value
			values[i].Value = float64(r.Count)
		}

		SuccessResponse(c, gin.H{
			"values":    values,
			"parameter": parameter,
		})
	}
}