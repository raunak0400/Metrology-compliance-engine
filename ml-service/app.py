"""
Placeholder Python ML service for the Legal Metrology Compliance Checker.

This is a STUB. It exists only so the Node backend can orchestrate the full
pipeline locally. The real service (OpenCV, PaddleOCR, custom object-detection
and NER/classification models) is owned by the ML team and lives separately.

It intentionally returns deterministic mock structures that match the contract
the Node `ml.client` expects. It does NOT perform any real computer vision.

Contract:
  GET  /health
  POST /ml/ocr             -> raw OCR blocks + bounding boxes + confidence
  POST /ml/detect-label    -> detected label / panel regions
  POST /ml/extract-fields  -> candidate declaration fields (MRP, net qty, ...)
  POST /ml/analyze         -> combined pipeline output
"""

from datetime import datetime, timezone
from typing import List, Optional

from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="Legal Metrology ML Service (placeholder)")

MODEL_VERSION = "stub-0.0.0"


class ImageRef(BaseModel):
    inspectionId: str
    imageId: str
    storageKey: str
    # A short-lived signed URL the ML service can download from.
    imageUrl: Optional[str] = None


class BBox(BaseModel):
    x: float
    y: float
    width: float
    height: float


class OcrBlock(BaseModel):
    text: str
    confidence: float
    bbox: BBox


@app.get("/health")
def health():
    return {"status": "ok", "modelVersion": MODEL_VERSION}


@app.post("/ml/ocr")
def ocr(ref: ImageRef):
    """Return raw OCR blocks. Real service uses PaddleOCR + preprocessing."""
    return {
        "inspectionId": ref.inspectionId,
        "imageId": ref.imageId,
        "modelVersion": MODEL_VERSION,
        "blocks": [
            {"text": "M.R.P. Rs. 199.00", "confidence": 0.94,
             "bbox": {"x": 40, "y": 120, "width": 220, "height": 30}},
            {"text": "Net Qty: 500 g", "confidence": 0.91,
             "bbox": {"x": 40, "y": 160, "width": 180, "height": 28}},
        ],
        "preprocessing": {"skewCorrected": True, "denoised": True},
        "createdAt": datetime.now(timezone.utc).isoformat(),
    }


@app.post("/ml/detect-label")
def detect_label(ref: ImageRef):
    """Return detected label / principal-display-panel regions."""
    return {
        "inspectionId": ref.inspectionId,
        "imageId": ref.imageId,
        "modelVersion": MODEL_VERSION,
        "regions": [
            {"type": "PRINCIPAL_DISPLAY_PANEL", "confidence": 0.88,
             "bbox": {"x": 20, "y": 100, "width": 300, "height": 400}},
        ],
        "createdAt": datetime.now(timezone.utc).isoformat(),
    }


@app.post("/ml/extract-fields")
def extract_fields(ref: ImageRef):
    """Return candidate declaration fields (NER / classification output)."""
    return {
        "inspectionId": ref.inspectionId,
        "imageId": ref.imageId,
        "modelVersion": MODEL_VERSION,
        "fields": [
            {"field": "MRP", "rawValue": "Rs. 199.00", "confidence": 0.93,
             "bbox": {"x": 40, "y": 120, "width": 220, "height": 30}},
            {"field": "NET_QUANTITY", "rawValue": "500 g", "confidence": 0.90,
             "bbox": {"x": 40, "y": 160, "width": 180, "height": 28}},
            {"field": "MANUFACTURER", "rawValue": "Example Foods Pvt Ltd, Industrial Area, Pune, India", "confidence": 0.85,
             "bbox": {"x": 40, "y": 200, "width": 260, "height": 26}},
            {"field": "CONSUMER_CARE", "rawValue": "care@example-foods.example, 1800-000-0000", "confidence": 0.82,
             "bbox": {"x": 40, "y": 230, "width": 260, "height": 24}},
            {"field": "MANUFACTURE_DATE", "rawValue": "01/2026", "confidence": 0.88,
             "bbox": {"x": 40, "y": 258, "width": 140, "height": 22}},
        ],
        "createdAt": datetime.now(timezone.utc).isoformat(),
    }


@app.post("/ml/analyze")
def analyze(ref: ImageRef):
    """Combined pipeline output: the consolidated analysis JSON contract.

    Real/OCR/detection/field-extraction sub-results plus the broader
    per-package analysis (category, package geometry, panel/text analysis,
    scale reference for pixel->mm conversion). Deterministic mock values only
    — no computer vision actually runs here. Never includes image bytes; the
    image is already in S3/MinIO (see ref.storageKey / ref.imageUrl).
    """
    return {
        "inspectionId": ref.inspectionId,
        "imageId": ref.imageId,
        "modelVersion": MODEL_VERSION,
        "ocr": ocr(ref),
        "regions": detect_label(ref)["regions"],
        "fields": extract_fields(ref)["fields"],
        "createdAt": datetime.now(timezone.utc).isoformat(),
        "productId": None,
        "category": "GENERAL",
        "saleChannel": "RETAIL",
        "netWeightKg": 0.5,
        "isImported": False,
        "packageShape": "RECTANGULAR_BOX",
        "packageSurfaceType": "PRINTED_PLASTIC_FILM",
        # A 10px/mm reference (e.g. a printed calibration card in frame) —
        # lets the Go measurement package convert declaration bbox heights to
        # real-world mm instead of reporting hasValidScale=false.
        "scaleReferenceMm": {"pxPerMm": 10.0, "source": "REFERENCE_CARD"},
        "panelMetadata": {"panelType": "PRINCIPAL_DISPLAY_PANEL", "areaCm2": 48.0},
        "panelAnalysis": {},
        "textOverlapAnalysis": {"overlapsDetected": False},
        "extractedTags": ["prepackaged", "retail-sale"],
        "textMeasurements": {},
        "extraFields": {},
    }
