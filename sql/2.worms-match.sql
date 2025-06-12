/*
 * WoRMS Matching Preparation Script
 * Purpose: Prepare organism data for WoRMS (World Register of Marine Species) matching
 * Author: MarineNP Team
 * Date: 2025-06-10
 * 
 * This script fixes IRIs in the organisms table and exports organisms
 * with missing IRIs for WoRMS matching.
 */

-- IRI Fixes
-- Fix URL-encoded characters in IRIs to make them properly readable
UPDATE public.organisms
SET iri = REPLACE(
            REPLACE(iri, '%3A%2F%2F', '://'),
            '%2F', '/'
          )
WHERE iri LIKE '%3A%2F%2F%' OR iri LIKE '%2F%';

-- Data Export
-- Export organisms with missing IRIs to CSV for WoRMS matching
-- This CSV will be used to match organisms with their WoRMS identifiers
COPY (SELECT * FROM "organisms" WHERE "iri" IS NULL ORDER BY "iri") TO STDOUT WITH CSV HEADER;

