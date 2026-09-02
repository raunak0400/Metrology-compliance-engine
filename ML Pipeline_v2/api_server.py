import os
import json
from pathlib import Path
from typing import List, Optional
from dotenv import load_dotenv
from fastapi import FastAPI, File, Form, UploadFile, HTTPException, status
from fastapi.responses import JSONResponse
from fastapi.staticfiles import StaticFiles

# Load environment variables from .env file automatically
load_dotenv()

# Import directly from your python verification file
from legal_metrology_label_verification_v3 import process_label_verification_multi

app = FastAPI(title="Legal Metrology Verification Engine")

# Path where processed images are written by process_label_verification_multi
BASE_DIR = Path(__file__).resolve().parent
PROCESSED_DIR = BASE_DIR / "processed_images"
PROCESSED_DIR.mkdir(parents=True, exist_ok=True)

# Expose processed_images for browser rendering
app.mount("/images", StaticFiles(directory=str(PROCESSED_DIR)), name="images")

@app.post("/api/v1/verify-label")
async def verify_label(
    files: List[UploadFile] = File(...),
    product_id: str = Form("SKU-001"),
    category: str = Form("other"),
    sale_channel: str = Form("retail"),
    net_weight_kg: float = Form(0.273),
    is_imported: bool = Form(False),
    package_surface_type: str = Form("printed"),
    physical_height_mm: float = Form(260.0),
    physical_width_mm: float = Form(130.0),
    api_key: Optional[str] = Form(None)
):
    return await _handle_verify(
        files=files,
        product_id=product_id,
        category=category,
        sale_channel=sale_channel,
        net_weight_kg=net_weight_kg,
        is_imported=is_imported,
        package_surface_type=package_surface_type,
        physical_height_mm=physical_height_mm,
        physical_width_mm=physical_width_mm,
        api_key=api_key,
    )


async def _handle_verify(
    files: List[UploadFile],
    product_id: str,
    category: str,
    sale_channel: str,
    net_weight_kg: float,
    is_imported: bool,
    package_surface_type: str,
    physical_height_mm: float,
    physical_width_mm: float,
    api_key: Optional[str] = None,
):
    if not files:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="No files uploaded."
        )

    try:
        image_bytes_list = []
        for file in files:
            content = await file.read()
            if content:
                image_bytes_list.append(content)

        if not image_bytes_list:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Uploaded file buffers are empty."
            )

        # Determine API key (form value takes precedence, then .env / environment)
        final_api_key = api_key or os.getenv("GEMINI_API_KEY")
        if not final_api_key:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="GEMINI_API_KEY is not set in environment or .env file."
            )

        # Execute processing pipeline directly
        result = process_label_verification_multi(
            image_bytes_list=image_bytes_list,
            api_key=final_api_key,
            product_id=product_id,
            category=category,
            sale_channel=sale_channel,
            net_weight_kg=net_weight_kg,
            is_imported=is_imported,
            package_surface_type=package_surface_type,
            physical_height_mm=physical_height_mm,
            physical_width_mm=physical_width_mm
        )

        # Save result JSON to current execution folder as per python script logic
        json_output_path = BASE_DIR / "label_verification_result.json"
        with open(json_output_path, "w", encoding="utf-8") as f:
            json.dump(result, f, indent=2, ensure_ascii=False)

        # Transform local file system paths into accessible FastAPI URLs
        web_annotated_paths = []
        for p_str in result.get("all_annotated_image_paths", []):
            file_name = Path(p_str).name
            web_annotated_paths.append(f"http://localhost:8000/images/{file_name}")

        result["all_annotated_image_urls"] = web_annotated_paths
        if web_annotated_paths:
            result["primary_image_url"] = web_annotated_paths[0]

        return JSONResponse(content=result, status_code=200)

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )


@app.post("/api/verify")
async def verify_label_legacy(
    files: List[UploadFile] = File(...),
    product_id: str = Form("SKU-001"),
    category: str = Form("other"),
    sale_channel: str = Form("retail"),
    net_weight_kg: float = Form(0.273),
    is_imported: bool = Form(False),
    package_surface_type: str = Form("printed"),
    physical_height_mm: float = Form(260.0),
    physical_width_mm: float = Form(130.0),
    api_key: Optional[str] = Form(None)
):
    """Compatibility endpoint for frontend code posting to `/api/verify`."""
    return await _handle_verify(
        files=files,
        product_id=product_id,
        category=category,
        sale_channel=sale_channel,
        net_weight_kg=net_weight_kg,
        is_imported=is_imported,
        package_surface_type=package_surface_type,
        physical_height_mm=physical_height_mm,
        physical_width_mm=physical_width_mm,
        api_key=api_key,
    )

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("api_server:app", host="0.0.0.0", port=8000, reload=True)