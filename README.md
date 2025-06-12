# MarineNP: A Comprehensive and Locally Deployable Database of Marine Natural Products

MarineNP is a powerful database application that provides access to comprehensive information about marine natural products. The application can be deployed locally and offers advanced search capabilities, detailed compound information, and interactive features.

## Features

- **Keyword Search**: Search across multiple fields including compound names, SMILES structures, identifiers, CAS numbers, and more
- **Advanced Search**: Powerful filtering options with over 40 searchable properties
- **Detailed Compound Information**: Access comprehensive data including:
  - Basic compound information
  - Structure information (InChI, SMILES)
  - Chemical classification
  - Physical properties
  - Biological sources
  - Geographic distribution data

## System Requirements

| Category | Requirement |
|----------|-------------|
| Memory | Minimum 8GB RAM |
| Storage | Minimum 32GB free disk space |
| Operating System | Windows 10 (x64) or newer, macOS (Apple Silicon), Linux (tested on Ubuntu 24.04) |
| Network | Internet access required only for initial file downloads |

## Quick Installation Guide

1. **Download Package**
   - Download the application package from [MarineNP Downloads](https://marinenp.scicloud.eu/#/data-access/downloads)

2. **Extract Files**
   - Extract the downloaded zip file to your preferred directory
   - The package includes all necessary static files, executables, and database files

3. **Run the Application**
   - **Windows**: Double-click `marinenp-windows.exe`
   - **Linux**: 
     ```bash
     chmod +x marinenp-linux
     ./marinenp-linux
     ```
   - **macOS**:
     ```bash
     chmod +x marinenp-macos
     ./marinenp-macos
     ```

4. **Access the Application**
   - Open your web browser and navigate to `http://localhost:3000`

## Advanced Configuration

### Custom Port
Create a `.env` file in the application directory with:
```plaintext
PORT=8080
```
Replace `8080` with your desired port number.

### Database Updates
To update the database:
1. Download the latest SQLite database file from the [Downloads](https://marinenp.scicloud.eu/#/data-access/downloads) page
2. Stop the running application
3. Replace the existing database file with the downloaded one
4. Restart the application

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Port already in use | Change the port in the .env file or stop the application using the current port |
| Application won't start | Ensure you have extracted all static files and they are in the correct location |
| Database errors | Download a fresh copy of the database file from the Downloads page |

## User Guides

For detailed information about using MarineNP, please refer to our [comprehensive user guides](https://marinenp.scicloud.eu/#/help/guides).

## Note
Internet access is only required for the initial download of files. Once all files are downloaded, the application can run offline.
