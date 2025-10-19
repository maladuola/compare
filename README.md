# Mogost Toolkit

Professional file processing and comparison utilities built for the Mogost team.

## Features

### Tool 1: File Comparison
- Upload two files to compare their content
- Supports side-by-side alignment similar to Beyond Compare
- Highlights differences in red
- Provides line-by-line comparison and diff analysis

### Tool 2: CSV Viewer
- Upload a CSV file and inspect its content
- Renders the data in a table-like layout
- Automatically detects data types and column structure
- Provides file statistics

### Tool 3: Archive Trade Comparison
- Upload a ZIP archive and extract it automatically
- Analyses directory structures (supports ABC, ABD, ABE, and similar folders)
- Automatically detects trade files (`babyy-risk-{id}.txt` and `candyy-risk-{id}.txt`)
- Compares paired files for the same trade ID
- Supports batch comparisons across multiple directories

## Tech Stack

- **Backend**: Go + Gin framework
- **Frontend**: HTML5 + CSS3 + JavaScript
- **File handling**: Supports multiple file formats
- **Diff algorithm**: Powered by the go-diff library

## Getting Started

### 1. Requirements
- Go 1.21 or newer
- Modern browser with HTML5 support

### 2. Build and Run

```bash
# Download dependencies
go get -v

# Build the binary
go build -v -o tools

# Run the server
./tools
```

Or use the build script:

```bash
# Make the script executable
chmod +x build.sh

# Run the build script
./build.sh

# Run the server
./tools
```

### 3. Access the tools

Open <http://localhost:8080> in your browser.

## Usage

### File Comparison
1. Click the "File Comparison" card
2. Upload the two files you want to compare
3. The server generates the diff automatically
4. Review the side-by-side output

### CSV Viewer
1. Click the "CSV Viewer" card
2. Upload a CSV file
3. Inspect the table view of the data
4. Review the file statistics

### Archive Trade Comparison
1. Click the "Archive Trade Comparison" card
2. Upload a ZIP archive
3. The server extracts the archive and analyses each directory
4. Review the diff for the detected trade files

## Project Structure

```
mogost-tools/
├── main.go                 # Application entry point
├── go.mod                  # Go module configuration
├── build.sh                # Build script
├── README.md               # Project documentation
├── tools/                  # Backend tooling package
│   ├── file_compare.go     # File comparison handlers
│   ├── csv_viewer.go       # CSV viewer handlers
│   └── archive_compare.go  # Archive comparison handlers
├── templates/              # Frontend templates
│   └── index.html          # Main page
├── uploads/                # Uploaded files
│   ├── file-compare/       # File comparison uploads
│   ├── csv/                # CSV uploads
│   └── archive-compare/    # Archive uploads
└── temp/                   # Temporary files
```

## API

### File Comparison
- `POST /api/file-compare/upload` - Upload files for comparison
- `GET /api/file-compare/compare` - Generate a diff between two files

### CSV Viewer
- `POST /api/csv/upload` - Upload a CSV file
- `GET /api/csv/view` - Retrieve CSV content and statistics

### Archive Comparison
- `POST /api/archive-compare/upload` - Upload a ZIP archive
- `GET /api/archive-compare/compare` - Compare detected trade files

## Development Notes

### Adding a New Tool
1. Create a new handler file in the `tools/` directory
2. Implement the required handler functions
3. Register the routes in `main.go`
4. Update the frontend with the new interface

### Customising Styles
- Edit the CSS inside `templates/index.html`
- Responsive layouts and modern UI styles are supported

## Notes

1. Uploaded files are stored under `uploads/`; clean them up regularly
2. Extracted archives temporarily occupy disk space and are cleaned after processing
3. Configure appropriate file size limits for production
4. Ensure the server has enough disk space for large files
 
 