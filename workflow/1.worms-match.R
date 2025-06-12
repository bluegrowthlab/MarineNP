# MarineNP WoRMS Matching
# Purpose: Match organism names to WoRMS (World Register of Marine Species)
# Author: MarineNP Team
# Date: 2025-06-10
#
# This file performs taxonomic name matching against the WoRMS database
# to standardize organism names and retrieve additional metadata.

# install.packages("worrms")
# install.packages("dplyr")

library(worrms)
library(dplyr)

# Read the TSV file into a dataframe, which is derived from the output of the SQL query
# in sql/worms-match.sql and matched to WoRMS using the LifeWatch service.
# https://www.lifewatch.be/e-lab/
coconut_worms <- read.delim(file = 'sql/worms_match_coconut_organisms.tsv', sep = '\t', header = T)

# Define the 'taxonmatch_note_worms' values that need API calls
fuzzy_categories <- c(
  "Empty response",
  "fuzzy matches : more than one possibility",
  "fuzzy matches(exact) : more than one possibility",
  "fuzzy matches(near_1) : more than one possibility",
  "fuzzy matches(near_2) : more than one possibility",
  "fuzzy matches(near_3) : more than one possibility",
  "fuzzy matches(phonetic) : more than one possibility"
)

# Identify rows that need updating
# Ensure taxonmatch_note_worms is a character vector for easier matching
coconut_worms$taxonmatch_note_worms <- as.character(coconut_worms$taxonmatch_note_worms)
rows_to_update_indices <- which(coconut_worms$taxonmatch_note_worms %in% fuzzy_categories)

# Print how many rows will be processed
print(paste("Number of rows to process:", length(rows_to_update_indices)))

# Loop through the identified rows and update them
# This loop can take a while if there are many rows, as it makes an API call for each.
for (i in rows_to_update_indices) {
  original_name <- coconut_worms$scientificname[i] # Or the column used for the initial WoRMS query
  
  print(paste("Processing row", which(rows_to_update_indices == i), "of", length(rows_to_update_indices), ":", original_name))
  
  tryCatch({
    # Fetch records using taxamatch (fuzzy matching)
    # wm_records_taxamatch returns a list, where each element is a dataframe of matches for a name
    # We are querying one name at a time here.
    # You might want to set marine_only = TRUE if appropriate, depending on your dataset.
    # The default for marine_only in wm_records_taxamatch is TRUE.
    matches_list <- wm_records_taxamatch(original_name) 
    
    # The result for a single name query is a list with one element, which is a dataframe
    if (length(matches_list) > 0 && !is.null(matches_list[[1]]) && nrow(matches_list[[1]]) > 0) {
      first_match <- matches_list[[1]][1, ] # Take the first row (first match)
      
      # Update the columns in your coconut_worms dataframe
      # Ensure the column names from 'first_match' correspond to what wm_records_taxamatch returns
      # Common columns from worrms results: AphiaID, scientificname, status, valid_AphiaID, valid_name
      
      coconut_worms$name_aphia_worms[i] <- ifelse(!is.null(first_match$scientificname), first_match$scientificname, NA)
      coconut_worms$aphiaid_worms[i] <- ifelse(!is.null(first_match$AphiaID), first_match$AphiaID, NA)
      coconut_worms$accepted_name_aphia_worms[i] <- ifelse(!is.null(first_match$valid_name), first_match$valid_name, NA)
      coconut_worms$valid_aphiaid_worms[i] <- ifelse(!is.null(first_match$valid_AphiaID), first_match$valid_AphiaID, NA)
      coconut_worms$status_aphia_worms[i] <- ifelse(!is.null(first_match$status), first_match$status, NA)
      
      # Optionally, update the note to reflect that it was resolved
      coconut_worms$taxonmatch_note_worms[i] <- paste0("fuzzy_resolved_to_first_match (was: ", coconut_worms$taxonmatch_note_worms[i], ")")
      
      # If the original match count was from the previous tool, you might want to set it to 1
      # coconut_worms$taxonmatch_matchcount_worms[i] <- 1 

    } else {
      print(paste("No matches found via API for:", original_name))
      # Optionally, update the note if no match is found on this attempt
      coconut_worms$taxonmatch_note_worms[i] <- paste0("fuzzy_API_no_match (was: ", coconut_worms$taxonmatch_note_worms[i], ")")
    }
  }, error = function(e) {
    print(paste("Error processing:", original_name, "-", e$message))
    # Optionally, update the note to reflect the error
    coconut_worms$taxonmatch_note_worms[i] <- paste0("fuzzy_API_error (was: ", coconut_worms$taxonmatch_note_worms[i], ")")
  })
  
  # Add a small delay to be polite to the API
  Sys.sleep(1) 
}

print("Processing complete.")

# Check the updated notes:
print(summary(as.factor(coconut_worms$taxonmatch_note_worms)))

# Filter the dataframe to only include rows where the environment_aphia_worms column
# contains "Marine" or "Brackish"
coconut_ogsm_marine <- coconut_worms %>% 
  filter(grepl("Marine", environment_aphia_worms) | grepl("Brackish", environment_aphia_worms)) %>%
  select(id, aphiaid_worms, name_aphia_worms, environment_aphia_worms)

# Write the filtered dataframe to a TSV file
write.csv(coconut_ogsm_marine, file = "sql/2.organisms_marine.csv",row.names = F)

