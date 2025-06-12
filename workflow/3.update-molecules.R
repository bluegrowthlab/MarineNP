# MarineNP Molecule Update
# Purpose: Update molecule data in the MarineNP database
# Author: MarineNP Team
# Date: 2025-06-10
#
# This file updates molecule records in the MarineNP database with additional
# chemical properties and metadata.

# Update the molecules table with marine status from organisms
# Rscript workflow/3.update-molecules.R

# Source the functions file
source("workflow/0.functions.R")

# Add is_marine column to molecules table
alter_query <- "
ALTER TABLE public.molecules
ADD COLUMN IF NOT EXISTS is_marine BOOLEAN DEFAULT FALSE;
"

# Execute the ALTER TABLE statement
execute_update(alter_query)

# Update molecules that are associated with marine organisms
update_query <- "
UPDATE public.molecules m
SET is_marine = TRUE
FROM public.molecule_organism mo
JOIN public.organisms o ON mo.organism_id = o.id
WHERE m.id = mo.molecule_id
AND o.is_marine = TRUE;
"

# Execute the update
execute_update(update_query)

# Create index on is_marine column
index_query <- "
CREATE INDEX IF NOT EXISTS idx_molecules_is_marine 
ON public.molecules(is_marine);
"

# Execute the index creation
execute_update(index_query)

print("Update complete!") 