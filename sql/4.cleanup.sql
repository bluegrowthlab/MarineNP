/*
 * Database Cleanup Script
 * Purpose: Remove non-marine data and unused tables from the MarineNP database
 * Author: MarineNP Team
 * Date: 2025-06-10
 * 
 * This script performs cleanup operations to maintain data quality by:
 * 1. Removing unused tables
 * 2. Deleting non-marine molecules and their related data
 * 3. Cleaning up the OBIS cache
 */

-- Table Cleanup
-- Drop tables that are no longer needed
DROP TABLE IF EXISTS tags CASCADE;
DROP TABLE IF EXISTS taggables CASCADE;
DROP TABLE IF EXISTS structures CASCADE;
DROP TABLE IF EXISTS molecule_related CASCADE;
DROP TABLE IF EXISTS collection_molecule CASCADE;
DROP TABLE IF EXISTS collections CASCADE;

-- Non-marine Data Cleanup
-- Create temporary table to store IDs of non-marine molecules for batch deletion
CREATE TEMPORARY TABLE temp_non_marine_molecules AS
SELECT id FROM molecules WHERE is_marine = false;

-- Delete Related Records
-- Remove all references to non-marine molecules from related tables
DELETE FROM citables 
WHERE citable_id IN (SELECT id FROM temp_non_marine_molecules);

DELETE FROM entries
USING temp_non_marine_molecules
WHERE entries.molecule_id = temp_non_marine_molecules.id;

DELETE FROM properties 
WHERE molecule_id IN (SELECT id FROM temp_non_marine_molecules);

DELETE FROM molecule_organism 
WHERE molecule_id IN (SELECT id FROM temp_non_marine_molecules);

DELETE FROM geo_location_molecule 
WHERE molecule_id IN (SELECT id FROM temp_non_marine_molecules);

-- Delete Non-marine Records
-- Remove non-marine molecules and organisms
DELETE FROM molecules
WHERE is_marine = false;

DELETE FROM organisms 
WHERE is_marine = false;

-- Cache Cleanup
-- Remove OBIS cache entries for organisms that no longer exist
DELETE FROM obis_cache 
WHERE aphiaid_worms NOT IN (
    SELECT aphiaid_worms 
    FROM organisms 
    WHERE aphiaid_worms IS NOT NULL
);

-- Cleanup
-- Drop temporary table used for batch operations
DROP TABLE temp_non_marine_molecules; 