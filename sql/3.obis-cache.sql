/*
 * OBIS Cache Setup Script
 * Purpose: Create and configure caching system for OBIS (Ocean Biogeographic Information System) data
 * Author: MarineNP Team
 * Date: 2025-06-10
 * 
 * This script sets up a caching mechanism for OBIS records to improve
 * performance and reduce API calls to the OBIS system.
 */

-- Table Creation
-- Create table for caching OBIS records with timestamps for tracking updates
CREATE TABLE IF NOT EXISTS obis_cache (
    id SERIAL PRIMARY KEY,
    aphiaid_worms INTEGER NOT NULL,
    obis_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(aphiaid_worms)
);

-- Index Creation
-- Create index on aphiaid_worms for faster lookups and joins
CREATE INDEX IF NOT EXISTS idx_obis_cache_aphiaid ON obis_cache(aphiaid_worms);

-- Trigger Function
-- Function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_obis_cache_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger Creation
-- Create trigger to maintain updated_at timestamp
CREATE TRIGGER update_obis_cache_updated_at
    BEFORE UPDATE ON obis_cache
    FOR EACH ROW
    EXECUTE FUNCTION update_obis_cache_updated_at(); 