# Legal Metrology Label Verification System

An automated AI-powered web application designed to process packaging photos, perform OCR and visual analysis, extract required Legal Metrology fields, and verify packaging compliance.

The platform combines a **Go Gateway Server**, a **FastAPI Python ML/OCR Service**, and a **Modern Single-Page Dashboard Interface**.

---

## Architecture Overview

```text
[ Browser Frontend ]
        │  (HTTP / Port 8080)
        ▼
[ Go Gateway Server ] (main.go)
        │  (Multipart Forwarding / Port 8000)
        ▼
[ FastAPI Backend Engine ] (api_server.py)
        │
        ├──► legal_metrology_label_verification_v3.py (Gemini AI & Analysis)
        └──► Saves annotated outputs to /processed_images
```

- **Frontend**: HTML5, CSS3, and JavaScript (Fetch API) for dynamic multi-image upload and real-time report rendering.
- **Gateway Server (Go)**: Listens on `:8080`, serves static web assets, handles CORS, and forwards multi-part requests to the Python engine.
- **Core Verification Engine (Python/FastAPI)**: Listens on `:8000`, communicates with the Gemini API, processes image bounding boxes, saves JSON results, and exposes `/images` for static previewing.

## Folder Structure
Plaintext

```
project_root/
│
├── .env                                       # Local environment variables (API Key)
├── README.md                                  # System documentation
├── main.go                                    # Go reverse proxy and web server
├── api_server.py                              # FastAPI REST API layer
├── legal_metrology_label_verification_v3.py  # Core verification logic and Gemini API pipeline
├── label_verification_result.json             # Output generated after execution
├── processed_images/                          # Output folder for annotated bounding box images
└── frontend/
    └── index.html                             # Single-page dashboard interface
```

## Prerequisites
Before running the application, ensure you have installed:

1. **Go** (`v1.18+`) — [Download Go](https://golang.org/dl/)
2. **Python** (`v3.9+`) — [Download Python](https://www.python.org/)
3. **Gemini API Key** from Google AI Studio.

## Environment Setup

1. **Clone the repository** (or navigate to your working directory):Bash
```
cd "path/to/your/project"
```
2. **Create a `.env` file** in the project root directory:Code snippet
```
GEMINI_API_KEY=your_actual_gemini_api_key_here
```
3. **Install Python dependencies**:Bash
```
pip install fastapi uvicorn python-dotenv google-generativeai pillow
```

## Running the Application
To run the complete system, start both the **FastAPI Python Backend** and the **Go Web Gateway**.

### Step 1: Start the FastAPI Service
In your first terminal window, run:

Bash

```
python api_server.py
```

- The API server will start on `http://localhost:8000`.

### Step 2: Start the Go Server
In a second terminal window, run:

Bash

```
go run main.go
```

- The Go Gateway will start on `http://localhost:8080`.

## How to Use

1. Open your web browser and navigate to:Plaintext
```
http://localhost:8080/
```
2. **Fill in Product Attributes**:Enter Product ID, Category, Sale Channel, and Net Weight.
3. Toggle the "Is Imported" option if applicable.
4. **Upload Package Images**:Click **Product Packaging Photos** and select one or multiple images (`.png`, `.jpg`, `.jpeg`).
5. *Tip*: Hold `Ctrl` (or `Cmd` on Mac) to select multiple files at once.
6. **Run Verification**:Click **Run Verification**.
7. View compliance status, annotated visual labels, and extracted fields rendered directly on the dashboard.

## API Endpoints

### 1. `POST /api/v1/verify-label` (FastAPI Engine)
Processes uploaded files and meta fields, executes Gemini verification, and outputs results.

- **Content-Type**: `multipart/form-data`
- **Form Parameters**:`files`: Image file binaries (Multiple files supported)
- `product_id` *(string)*: Unique SKU identifier
- `category` *(string)*: Product classification
- `sale_channel` *(string)*: Retail or E-commerce
- `net_weight_kg` *(float)*: Net weight value
- `is_imported` *(boolean)*: Import status flag

- **Response Example (`200 OK`)**:JSON
```
{
  "product_id": "SKU-001",
  "package_shape": "rectangular",
  "panel_analysis": {
    "COMPLIANT": true
  },
  "all_annotated_image_urls": [
    "http://localhost:8000/images/annotated_1.jpg"
  ],
  "extracted_fields": [
    {
      "field_type": "Net Quantity",
      "text": "273 g",
      "font_height_mm": 3.2
    }
  ]
}
```

### 2. `POST /api/verify` (Go Gateway)
Proxies incoming client requests directly to `http://localhost:8000/api/v1/verify-label`.

## Output Files

- **`label_verification_result.json`**: Generated automatically in the root folder after each execution containing raw OCR, extracted bounding boxes, and compliance evaluation.
- **`processed_images/`**: Contains visually annotated images highlighting detected label components.

## Troubleshooting

- **Go Command Not Found**: Ensure `C:\Go\bin` (or equivalent Go binary path) is added to your system `PATH` environment variable and restart your terminal.
- **Missing API Key Error**: Double-check that your `.env` file exists in the root directory and contains a valid `GEMINI_API_KEY`.
- **CORS / Upload Issues**: Ensure both FastAPI (`:8000`) and Go (`:8080`) are actively running concurrently.

