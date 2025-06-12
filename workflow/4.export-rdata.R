# MarineNP R Data Export
# Purpose: Export MarineNP data to R-compatible formats
# Author: MarineNP Team
# Date: 2025-06-10
#
# This file exports processed MarineNP data to R-compatible formats
# for further analysis and visualization.

csv_files <- list.files(pattern = "\\.csv$")

if (length(csv_files) == 0) {
  print("No .csv files were found in the current working directory.")
} else {
  loaded_df_names <- c()
  
  for (file_name in csv_files) {
    df_name <- sub("\\.csv$", "", file_name)
    data <- read.csv(file_name)
    assign(df_name, data)
    loaded_df_names <- c(loaded_df_names, df_name)
    cat("File '", file_name, "' has been loaded as data frame '", df_name, "'.\n", sep = "")
  }
  
  output_file <- "all_csv_data.RData"
  save(list = loaded_df_names, file = output_file)
  
  cat("\nAll data frames have been saved to '", output_file, "'.\n", sep = "")
  
  rm(csv_files, file_name, df_name, data, loaded_df_names, output_file)
  
  print("All .csv files have been successfully loaded and saved.")
}