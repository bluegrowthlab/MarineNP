/*
 * MarineNP Database Setup Script
 * Purpose: Initial database creation and user setup for MarineNP
 * Author: MarineNP Team
 * Date: 2025-06-10
 * 
 * This script creates the database and user with appropriate permissions
 * for the MarineNP project, which focuses on marine natural products.
 */

-- Database Creation
CREATE DATABASE marinenp;

-- User Setup
-- Create the user with a password (replace 'your_strong_password' with a secure password)
CREATE USER marinenp WITH PASSWORD 'your_strong_password';

-- Permission Configuration
-- Grant all privileges on the marinenp database to the marinenp user
GRANT ALL PRIVILEGES ON DATABASE marinenp TO marinenp;

-- Schema Permissions
-- Grant all privileges on all tables in the public schema
GRANT USAGE, CREATE ON SCHEMA public TO marinenp;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO marinenp;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO marinenp;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO marinenp;

/*
 * Data Import Instructions
 * The following commands should be run in the shell to import the COCONUT database:
 * 1. Download the database dump:
 *    wget https://zenodo.org/records/13897048/files/coconut-dump-10-2024.sql.zip
 * 2. Extract the dump file:
 *    unzip coconut-dump-10-2024.sql.zip
 * 3. Import the data:
 *    pg_restore -h localhost -U marinenp -d marinenp -W --no-owner --exit-on-error coconut-dump-10-2024.sql
 */