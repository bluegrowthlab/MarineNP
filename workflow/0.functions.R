# MarineNP Workflow Functions
# Purpose: Utility functions for the MarineNP data processing workflow
# Author: MarineNP Team
# Date: 2025-06-10
#
# This file contains helper functions used throughout the MarineNP data processing
# workflow, including data validation, transformation, and error handling.

# Load required packages
if (!require("RPostgres")) install.packages("RPostgres")
if (!require("dotenv")) install.packages("dotenv")

library(RPostgres)
library(dotenv)

#' Load environment variables from .env file
#' @return A list of environment variables
load_env <- function() {
  # Load .env file if it exists
  dotenv::load_dot_env()
  
  # Get environment variables with defaults matching config.go
  env_vars <- list(
    DB_HOST = Sys.getenv("DB_HOST", "localhost"),
    DB_PORT = as.numeric(Sys.getenv("DB_PORT", "5432")),
    DB_USER = Sys.getenv("DB_USER", "postgres"),
    DB_PASSWORD = Sys.getenv("DB_PASSWORD", "postgres"),
    DB_NAME = Sys.getenv("DB_NAME", "coconut"),
    DB_SSL_MODE = Sys.getenv("DB_SSL_MODE", "disable")
  )
  
  return(env_vars)
}

#' Execute a PostgreSQL query and return results as a data frame
#' @param query SQL query to execute
#' @param params Optional list of parameters for prepared statements
#' @return Data frame containing query results
#' @export
execute_query <- function(query, params = NULL) {
  # Load environment variables
  env <- load_env()
  
  # Create connection string
  con <- dbConnect(
    Postgres(),
    host = env$DB_HOST,
    port = env$DB_PORT,
    user = env$DB_USER,
    password = env$DB_PASSWORD,
    dbname = env$DB_NAME,
    sslmode = env$DB_SSL_MODE
  )
  
  # Ensure connection is closed when function exits
  on.exit(dbDisconnect(con))
  
  # Execute query
  if (is.null(params)) {
    result <- dbGetQuery(con, query)
  } else {
    result <- dbGetQuery(con, query, params)
  }
  
  return(result)
}

#' Execute a PostgreSQL query that doesn't return results
#' @param query SQL query to execute
#' @param params Optional list of parameters for prepared statements
#' @return TRUE if successful
#' @export
execute_update <- function(query, params = NULL) {
  # Load environment variables
  env <- load_env()
  
  # Create connection string
  con <- dbConnect(
    Postgres(),
    host = env$DB_HOST,
    port = env$DB_PORT,
    user = env$DB_USER,
    password = env$DB_PASSWORD,
    dbname = env$DB_NAME,
    sslmode = env$DB_SSL_MODE
  )
  
  # Ensure connection is closed when function exits
  on.exit(dbDisconnect(con))
  
  # Execute query
  if (is.null(params)) {
    dbExecute(con, query)
  } else {
    dbExecute(con, query, params)
  }
  
  return(TRUE)
} 