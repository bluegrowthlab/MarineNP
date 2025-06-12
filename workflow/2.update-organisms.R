# MarineNP Organism Update
# Purpose: Update organism data in the MarineNP database
# Author: MarineNP Team
# Date: 2025-06-10
#
# This file updates organism records in the MarineNP database with standardized
# taxonomic information and metadata from WoRMS.

# Update the organisms table with the WoRMS data
# Rscript workflow/2.update-organisms.R

# Source the functions file
source("workflow/0.functions.R")

# Add new columns to organisms table
alter_query <- "
ALTER TABLE public.organisms
ADD COLUMN IF NOT EXISTS aphiaid_worms INTEGER,
ADD COLUMN IF NOT EXISTS name_aphia_worms VARCHAR(255),
ADD COLUMN IF NOT EXISTS environment_aphia_worms TEXT,
ADD COLUMN IF NOT EXISTS is_marine BOOLEAN DEFAULT FALSE;
"

# Execute the ALTER TABLE statement
execute_update(alter_query)

# Read the marine organisms data
marine_organisms <- read.csv("sql/2.organisms_marine.csv")

# Process in batches of 25
batch_size <- 25
total_rows <- nrow(marine_organisms)
num_batches <- ceiling(total_rows / batch_size)

for (batch in 1:num_batches) {
  start_idx <- (batch - 1) * batch_size + 1
  end_idx <- min(batch * batch_size, total_rows)
  
  # Create the batch update query
  batch_query <- "
  UPDATE public.organisms o
  SET aphiaid_worms = c.aphiaid_worms,
      name_aphia_worms = c.name_aphia_worms,
      environment_aphia_worms = c.environment_aphia_worms,
      is_marine = TRUE
  FROM (VALUES "
  
  # Add values for each row in the batch
  values_list <- list()
  for (i in start_idx:end_idx) {
    values_list <- c(values_list, sprintf(
      "(%d, %d, '%s', '%s')",
      marine_organisms$id[i],
      marine_organisms$aphiaid_worms[i],
      gsub("'", "''", marine_organisms$name_aphia_worms[i]),  # Escape single quotes
      gsub("'", "''", marine_organisms$environment_aphia_worms[i])  # Escape single quotes
    ))
  }
  
  batch_query <- paste0(batch_query, paste(values_list, collapse = ","), ") AS c(id, aphiaid_worms, name_aphia_worms, environment_aphia_worms)")
  batch_query <- paste0(batch_query, " WHERE o.id = c.id;")
  
  # Execute the batch update
  execute_update(batch_query)
  
  print(sprintf("Processed batch %d of %d (rows %d to %d)", 
                batch, num_batches, start_idx, end_idx))
}

# Create index on is_marine column
index_query <- "
CREATE INDEX IF NOT EXISTS idx_organisms_is_marine 
ON public.organisms(is_marine);
"

# Execute the index creation
execute_update(index_query)

print("Update complete!") 