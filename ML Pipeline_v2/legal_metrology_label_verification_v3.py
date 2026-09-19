"""
LEGAL METROLOGY LABEL VERIFICATION - MULTI-PHOTO VERSION

Design goals:
- Multi-image horizontal stitching (handles 2, 3, or more photos of the same product)
- Gemini 3.6 Flash extraction of 44 Legal Metrology fields
- PaddleOCR bounding boxes executing in an isolated sub-process (Windows-safe with timeout)
- Strict compliance with requested output JSON schema

Required:
    pip install opencv-python numpy google-genai

Optional OCR:
    pip install paddleocr paddlepaddle

Gemini API key setup:
PowerShell:
    $env:GEMINI_API_KEY="YOUR_API_KEY"

Run:
    python legal_metrology_label_verification_v3.py
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple
from dotenv import load_dotenv

os.environ.setdefault("FLAGS_use_mkldnn", "0")
os.environ.setdefault("PADDLE_PDX_ENABLE_MKLDNN_BYDEFAULT", "0")
os.environ.setdefault("DISABLE_MODEL_SOURCE_CHECK", "True")

# Load environment variables from .env file automatically
load_dotenv()

import cv2
import numpy as np

from google import genai
from google.genai import types


# ================================================================
# 44 LEGAL METROLOGY FIELDS
# ================================================================

TAG_TO_FIELD_TYPE_MAP = {
    "MANUFACTURER_NAME": "manufacturer",
    "MANUFACTURER_ADDRESS": "manufacturer",
    "MANUFACTURER_EMAIL": "manufacturer",
    "MANUFACTURER_PHONE": "manufacturer",
    "MANUFACTURER_STATE": "manufacturer",
    "MANUFACTURER_PINCODE": "manufacturer",
    "PACKER_NAME": "packer",
    "PACKER_ADDRESS": "packer",
    "PACKER_EMAIL": "packer",
    "PACKER_PHONE": "packer",
    "IMPORTER_NAME": "importer",
    "IMPORTER_ADDRESS": "importer",
    "IMPORTER_EMAIL": "importer",
    "IMPORTER_PHONE": "importer",
    "COUNTRY_OF_ORIGIN": "country_of_origin",
    "PRODUCT_COMMON_NAME": "generic_name",
    "BRAND_NAME": "brand_name",
    "MODEL_NUMBER": "model_number",
    "SERIAL_NUMBER": "serial_number",
    "DIMENSIONS_LENGTH": "dimensions",
    "DIMENSIONS_WIDTH": "dimensions",
    "DIMENSIONS_HEIGHT": "dimensions",
    "NET_QUANTITY": "net_quantity",
    "NET_QUANTITY_UNIT": "net_quantity",
    "NUMBER_OF_UNITS": "unit_count",
    "MRP": "mrp",
    "MRP_INCLUSIVE_OF_TAXES": "mrp",
    "MRP_IN_INDIAN_CURRENCY": "mrp",
    "UNIT_SALE_PRICE": "unit_sale_price",
    "MANUFACTURE_MONTH": "mfg_date",
    "MANUFACTURE_YEAR": "mfg_date",
    "IMPORT_MONTH": "import_date",
    "IMPORT_YEAR": "import_date",
    "EXPIRY_DATE": "expiry_date",
    "BEST_BEFORE": "best_before",
    "USE_BY_DATE": "use_by_date",
    "BATCH_OR_LOT_NUMBER": "batch_number",
    "FSSAI_LICENSE_NUMBER": "fssai_license",
    "CONSUMER_CARE_NAME": "consumer_care",
    "CONSUMER_CARE_ADDRESS": "consumer_care",
    "CONSUMER_CARE_PHONE": "consumer_care",
    "CONSUMER_CARE_EMAIL": "consumer_care",
    "INGREDIENTS_LIST": "ingredients",
    "WARNING_AND_SAFETY_INSTRUCTIONS": "safety_instructions",
}

ALL_FIELDS = list(TAG_TO_FIELD_TYPE_MAP.keys())
PACKAGE_SHAPES = {
    "CUBOID_3D",
    "CYLINDRICAL",
    "FLAT_POUCH",
    "RECTANGULAR_PRISM",
}


# ================================================================
# GENERAL HELPERS
# ================================================================

def log(message: str) -> None:
    print(message, flush=True)


def safe_float(value: Any, default: float = 0.0) -> float:
    try:
        return float(value)
    except (TypeError, ValueError):
        return default


def clean_json_text(text: str) -> str:
    text = (text or "").strip()
    text = re.sub(r"^\s*```(?:json)?\s*", "", text, flags=re.I)
    text = re.sub(r"\s*```\s*$", "", text)
    return text.strip()


def normalize_bbox(value: Any) -> List[List[float]]:
    if value is None:
        return []

    if hasattr(value, "tolist"):
        value = value.tolist()

    try:
        arr = np.asarray(value, dtype=float)
        if arr.ndim == 1 and arr.size == 4:
            x1, y1, x2, y2 = arr.tolist()
            return [[x1, y1], [x2, y1], [x2, y2], [x1, y2]]

        if arr.ndim == 2 and arr.shape[1] >= 2:
            return [[float(point[0]), float(point[1])] for point in arr]
    except Exception:
        pass

    return []


def bbox_from_polygon(poly: List[List[float]]) -> Dict[str, int]:
    if not poly:
        return {"x": 0, "y": 0, "w": 0, "h": 0}

    xs = [p[0] for p in poly]
    ys = [p[1] for p in poly]

    x1 = int(round(min(xs)))
    y1 = int(round(min(ys)))
    x2 = int(round(max(xs)))
    y2 = int(round(max(ys)))

    return {
        "x": max(0, x1),
        "y": max(0, y1),
        "w": max(0, x2 - x1),
        "h": max(0, y2 - y1),
    }


# ================================================================
# IMAGE FUNCTIONS
# ================================================================

def read_image_file(path: Path) -> Optional[np.ndarray]:
    try:
        data = path.read_bytes()
        if not data:
            return None
        arr = np.frombuffer(data, dtype=np.uint8)
        return cv2.imdecode(arr, cv2.IMREAD_COLOR)
    except Exception:
        return None


def stitch_images(images: List[np.ndarray]) -> np.ndarray:
    valid = [img for img in images if img is not None and img.size > 0]

    if not valid:
        raise ValueError("No valid images were supplied.")

    if len(valid) == 1:
        return valid[0].copy()

    min_height = min(img.shape[0] for img in valid)
    resized = []

    for img in valid:
        h, w = img.shape[:2]
        new_width = max(1, int(round(w * min_height / h)))
        resized.append(cv2.resize(img, (new_width, min_height), interpolation=cv2.INTER_AREA))

    return cv2.hconcat(resized)


def save_image(path: Path, image: np.ndarray) -> None:
    ok, encoded = cv2.imencode(".jpg", image, [int(cv2.IMWRITE_JPEG_QUALITY), 92])
    if not ok:
        raise RuntimeError(f"Could not encode image: {path}")
    path.write_bytes(encoded.tobytes())


def draw_mandatory_field_boxes(
    image: np.ndarray, extracted_fields: Any
) -> List[np.ndarray]:
    """Dynamically scales label text to image resolution and ignores empty image panels."""
    if not isinstance(extracted_fields, list):
        return []

    field_title_map = {
        # Mandatory 7
        "manufacturer": "Name and Address",
        "packer": "Name and Address",
        "importer": "Name and Address",
        "country_of_origin": "Country of Origin",
        "generic_name": "Generic Name",
        "net_quantity": "Net Quantity",
        "mrp": "MRP",
        "mfg_date": "Manufacturing Date",
        "consumer_care": "Consumer Care",
        # Secondary / Other
        "brand_name": "Brand Name",
        "ingredients": "Ingredients",
        "fssai_license": "FSSAI License",
        "batch_number": "Batch No",
        "expiry_date": "Expiry Date",
        "best_before": "Expiry Date",
        "use_by_date": "Expiry Date",
        "unit_sale_price": "Unit Sale Price",
    }

    mandatory_7_titles = {
        "Name and Address",
        "Country of Origin",
        "Generic Name",
        "Net Quantity",
        "MRP",
        "Manufacturing Date",
        "Consumer Care",
    }

    img_h, img_w = image.shape[:2]
    parsed_items = []

    for record in extracted_fields:
        if not isinstance(record, dict):
            continue

        field_type = str(record.get("field_type", "")).lower()
        title = field_title_map.get(field_type, field_type.upper().replace("_", " "))

        bbox = record.get("bbox", {})
        x1, y1 = int(bbox.get("x", 0)), int(bbox.get("y", 0))
        w, h = int(bbox.get("w", 0)), int(bbox.get("h", 0))

        if w <= 0 or h <= 0:
            continue

        x2 = min(img_w - 1, x1 + w)
        y2 = min(img_h - 1, y1 + h)

        parsed_items.append({
            "title": title,
            "box": (x1, y1, x2, y2)
        })

    def boxes_are_close(b1, b2, margin=15):
        return not (b1[2] + margin < b2[0] or b1[0] - margin > b2[2] or 
                    b1[3] + margin < b2[1] or b1[1] - margin > b2[3])

    def merge_overlapping_items(items):
        if not items:
            return []
        
        clusters = []
        for item in items:
            placed = False
            for cluster in clusters:
                if any(boxes_are_close(item["box"], c["box"]) for c in cluster):
                    cluster.append(item)
                    placed = True
                    break
            if not placed:
                clusters.append([item])

        merged_result = []
        for cluster in clusters:
            min_x1 = min(c["box"][0] for c in cluster)
            min_y1 = min(c["box"][1] for c in cluster)
            max_x2 = max(c["box"][2] for c in cluster)
            max_y2 = max(c["box"][3] for c in cluster)
            
            titles = sorted(list(set(c["title"] for c in cluster)))
            combined_title = " / ".join(titles)

            merged_result.append({
                "title": combined_title,
                "box": (min_x1, min_y1, max_x2, max_y2)
            })
        return merged_result

    def render_layer(items):
        if not items:
            return None  # Skip returning images with no boxes

        merged_items = merge_overlapping_items(items)
        
        # 1. Determine bounding crop containing active boxes to remove empty side photos
        min_crop_x = min(item["box"][0] for item in merged_items)
        max_crop_x = max(item["box"][2] for item in merged_items)
        
        # Add side margins around detected group
        crop_x1 = max(0, min_crop_x - 40)
        crop_x2 = min(img_w, max_crop_x + 40)
        
        cropped_canvas = image[:, crop_x1:crop_x2].copy()
        crop_w = cropped_canvas.shape[1]

        # 2. Dynamic Text & Box Sizing relative to image resolution
        font_scale = max(0.45, crop_w / 1000.0)
        thickness = max(1, int(crop_w / 600.0))

        for item in merged_items:
            # Re-offset x coordinates relative to cropped view
            x1 = item["box"][0] - crop_x1
            y1 = item["box"][1]
            x2 = item["box"][2] - crop_x1
            y2 = item["box"][3]

            label_text = f" [{item['title']}] "

            # Thin green target box
            cv2.rectangle(cropped_canvas, (x1, y1), (x2, y2), (0, 255, 0), thickness)

            # High-visibility filled header box
            (txt_w, txt_h), baseline = cv2.getTextSize(
                label_text, cv2.FONT_HERSHEY_SIMPLEX, font_scale, thickness
            )

            top_y = max(0, y1 - txt_h - 8)
            bot_y = max(txt_h + 4, y1)
            right_x = min(crop_w - 1, x1 + txt_w + 6)

            cv2.rectangle(cropped_canvas, (x1, top_y), (right_x, bot_y), (0, 255, 0), -1)
            cv2.putText(
                cropped_canvas,
                label_text,
                (x1 + 2, bot_y - 4),
                cv2.FONT_HERSHEY_SIMPLEX,
                font_scale,
                (0, 0, 0),
                thickness,
                cv2.LINE_AA,
            )
        return cropped_canvas

    mandatory_group = [item for item in parsed_items if item["title"] in mandatory_7_titles]
    other_group = [item for item in parsed_items if item["title"] not in mandatory_7_titles]

    output_images = []
    
    img1 = render_layer(mandatory_group)
    if img1 is not None:
        output_images.append(img1)
        
    img2 = render_layer(other_group)
    if img2 is not None:
        output_images.append(img2)

    return output_images if output_images else [image.copy()]


# ================================================================
# STAGE 2 - PACKAGE / PDP
# ================================================================

class Stage2Preprocessor:

    def analyze_package_shape(self, image: np.ndarray) -> str:
        if image is None or image.size == 0:
            return "RECTANGULAR_PRISM"

        try:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
            blurred = cv2.GaussianBlur(gray, (5, 5), 0)
            edges = cv2.Canny(blurred, 50, 150)

            contours, _ = cv2.findContours(edges, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)

            if not contours:
                return "RECTANGULAR_PRISM"

            contour = max(contours, key=cv2.contourArea)
            if cv2.contourArea(contour) <= 0:
                return "RECTANGULAR_PRISM"

            perimeter = cv2.arcLength(contour, True)
            if perimeter <= 0:
                return "RECTANGULAR_PRISM"

            approx = cv2.approxPolyDP(contour, 0.02 * perimeter, True)
            _, _, width, height = cv2.boundingRect(contour)
            ratio = float(width) / float(height) if height > 0 else 1.0

            if len(approx) > 8 and 0.3 < ratio < 1.5:
                return "CYLINDRICAL"

            if len(approx) in (4, 5, 6):
                return "CUBOID_3D"

        except Exception:
            pass

        return "RECTANGULAR_PRISM"

    def extract_pdp(self, image: np.ndarray, package_shape: str) -> Tuple[np.ndarray, Dict[str, Any]]:
        height, width = image.shape[:2]

        # Preserve the full canvas after multiple package images are stitched.
        pdp = image.copy()
        ratio = 0.40 if package_shape == "CYLINDRICAL" else 1.0

        metadata = {
            "pdp_width_px": int(pdp.shape[1]),
            "pdp_height_px": int(pdp.shape[0]),
            "total_width_px": int(width),
            "total_height_px": int(height),
            "package_shape": package_shape,
            "pdp_area_ratio": ratio,
        }

        return pdp, metadata

    def calculate_required_pdp_area(self, package_shape: str, height_mm: float, width_mm: float) -> Tuple[float, float]:
        height_mm = max(0.0, safe_float(height_mm))
        width_mm = max(0.0, safe_float(width_mm))

        if package_shape == "CYLINDRICAL":
            total = height_mm * np.pi * width_mm
            required = total * 0.40
        else:
            total = height_mm * width_mm
            required = total

        return round(total, 2), round(required, 2)


# ================================================================
# STAGE 3 - GEMINI
# ================================================================

class GeminiExtractor:

    def __init__(self, api_key: Optional[str] = None):
        self.api_key = api_key or os.getenv("GEMINI_API_KEY")

        if not self.api_key:
            raise ValueError(
                "GEMINI_API_KEY is missing.\n\n"
                "In PowerShell run:\n"
                '$env:GEMINI_API_KEY="YOUR_API_KEY"\n\n'
                "Then run this Python file again."
            )

        self.client = genai.Client(api_key=self.api_key)

    def extract_tags_and_package_shape(
        self, image: np.ndarray
    ) -> Tuple[Dict[str, Any], Optional[str]]:
        ok, buffer = cv2.imencode(".jpg", image, [int(cv2.IMWRITE_JPEG_QUALITY), 95])
        if not ok:
            raise RuntimeError("Could not encode image for Gemini.")

        prompt = f"""
Analyze the provided product packaging image and perform Legal Metrology compliance extraction.

1. Classify the package 3D geometry as exactly one of: CUBOID_3D, CYLINDRICAL, FLAT_POUCH, RECTANGULAR_PRISM.
2. For each visible Legal Metrology field below, extract:
   - "text": The string printed on the package (or null if not visible).
   - "box_2d": The 2D bounding box as [ymin, xmin, ymax, xmax] using normalized coordinates from 0 to 1000. Return null if field is missing.

Fields:
MANUFACTURER_NAME, MANUFACTURER_ADDRESS, MANUFACTURER_EMAIL, MANUFACTURER_PHONE, MANUFACTURER_STATE, MANUFACTURER_PINCODE,
PACKER_NAME, PACKER_ADDRESS, PACKER_EMAIL, PACKER_PHONE, IMPORTER_NAME, IMPORTER_ADDRESS, IMPORTER_EMAIL, IMPORTER_PHONE,
COUNTRY_OF_ORIGIN, PRODUCT_COMMON_NAME, BRAND_NAME, MODEL_NUMBER, SERIAL_NUMBER, DIMENSIONS_LENGTH, DIMENSIONS_WIDTH,
DIMENSIONS_HEIGHT, NET_QUANTITY, NET_QUANTITY_UNIT, NUMBER_OF_UNITS, MRP, MRP_INCLUSIVE_OF_TAXES, MRP_IN_INDIAN_CURRENCY,
UNIT_SALE_PRICE, MANUFACTURE_MONTH, MANUFACTURE_YEAR, IMPORT_MONTH, IMPORT_YEAR, EXPIRY_DATE, BEST_BEFORE, USE_BY_DATE,
BATCH_OR_LOT_NUMBER, FSSAI_LICENSE_NUMBER, CONSUMER_CARE_NAME, CONSUMER_CARE_ADDRESS, CONSUMER_CARE_PHONE, CONSUMER_CARE_EMAIL,
INGREDIENTS_LIST, WARNING_AND_SAFETY_INSTRUCTIONS

Return ONLY valid JSON in this structure:
{{
  "PACKAGE_SHAPE": "CYLINDRICAL",
  "fields": {{
     "MANUFACTURER_NAME": {{"text": "...", "box_2d": [ymin, xmin, ymax, xmax]}},
     ...
  }}
}}
"""

        response = self.client.models.generate_content(
            model="gemini-3.6-flash",
            contents=[
                types.Part.from_bytes(data=buffer.tobytes(), mime_type="image/jpeg"),
                prompt,
            ],
            config=types.GenerateContentConfig(
                response_mime_type="application/json",
                temperature=0.1,
            ),
        )

        text = getattr(response, "text", None)
        if not text:
            raise RuntimeError("Gemini returned an empty response.")

        try:
            data = json.loads(clean_json_text(text))
        except json.JSONDecodeError as exc:
            raise RuntimeError("Gemini returned invalid JSON.") from exc

        if not isinstance(data, dict):
            raise RuntimeError("Gemini response is not a JSON object.")

        package_shape = data.get("PACKAGE_SHAPE")
        if package_shape not in PACKAGE_SHAPES:
            package_shape = None

        fields_data = data.get("fields", {})
        if not isinstance(fields_data, dict):
            fields_data = {}

        normalized_fields: Dict[str, Any] = {}
        for field in ALL_FIELDS:
            field_value = fields_data.get(field)
            if field_value is None:
                normalized_fields[field] = {"text": None, "box_2d": None}
                continue

            if isinstance(field_value, dict):
                text_value = field_value.get("text")
                box_value = field_value.get("box_2d")
            else:
                text_value = field_value
                box_value = None

            if isinstance(box_value, (list, tuple)) and len(box_value) == 4:
                try:
                    box_2d = [float(box_value[0]), float(box_value[1]), float(box_value[2]), float(box_value[3])]
                except (TypeError, ValueError):
                    box_2d = None
            else:
                box_2d = None

            normalized_fields[field] = {
                "text": str(text_value) if text_value is not None else None,
                "box_2d": box_2d,
            }

        if not normalized_fields:
            for field in ALL_FIELDS:
                value = data.get(field)
                if value is None:
                    normalized_fields[field] = {"text": None, "box_2d": None}
                    continue

                if isinstance(value, dict):
                    text_value = value.get("text")
                    box_value = value.get("box_2d")
                else:
                    text_value = value
                    box_value = None

                normalized_fields[field] = {
                    "text": str(text_value) if text_value is not None else None,
                    "box_2d": box_value if isinstance(box_value, (list, tuple)) and len(box_value) == 4 else None,
                }

        return normalized_fields, package_shape

    def extract_tags(self, image: np.ndarray) -> Dict[str, Any]:
        tags, _ = self.extract_tags_and_package_shape(image)
        return tags


# ================================================================
# ISOLATED PADDLEOCR WORKER
# ================================================================

def run_paddleocr_worker(image_path: Path, output_path: Path) -> int:
    try:
        from paddleocr import PaddleOCR

        image = read_image_file(image_path)
        if image is None:
            raise RuntimeError("OCR worker could not read temporary image.")

        try:
            ocr = PaddleOCR(
                lang="en",
                use_doc_orientation_classify=False,
                use_doc_unwarping=False,
                use_textline_orientation=False,
                enable_mkldnn=False,
            )
        except TypeError:
            ocr = PaddleOCR(lang="en", use_angle_cls=False, enable_mkldnn=False)

        if hasattr(ocr, "predict"):
            results = ocr.predict(image)
        else:
            results = ocr.ocr(image)

        blocks: List[Dict[str, Any]] = []

        if results is None:
            output_path.write_text(json.dumps([]), encoding="utf-8")
            return 0

        if not isinstance(results, (list, tuple)):
            results = [results]

        for result in results:
            data = None
            if isinstance(result, dict):
                data = result
            elif hasattr(result, "json"):
                try:
                    raw = result.json
                    data = json.loads(raw) if isinstance(raw, str) else raw
                except Exception:
                    data = None

            if isinstance(data, dict):
                polys = data.get("rec_polys") or data.get("dt_polys") or data.get("rec_boxes") or []
                texts = data.get("rec_texts") or data.get("rec_text") or []
                scores = data.get("rec_scores") or data.get("rec_score") or []

                if isinstance(texts, str):
                    texts = [texts]

                if np.isscalar(scores):
                    scores = [scores]

                for i, polygon in enumerate(polys):
                    poly = normalize_bbox(polygon)
                    text = str(texts[i]) if i < len(texts) else ""
                    score = safe_float(scores[i]) if i < len(scores) else 0.0

                    blocks.append({
                        "id": len(blocks),
                        "text": text,
                        "bbox_poly": poly,
                        "bbox": bbox_from_polygon(poly),
                        "ocr_confidence": round(score, 4),
                    })
                continue

            if isinstance(result, list):
                for line in result:
                    try:
                        if not isinstance(line, (list, tuple)) or len(line) < 2:
                            continue

                        polygon = normalize_bbox(line[0])
                        recognition = line[1]

                        if isinstance(recognition, (list, tuple)) and len(recognition) >= 2:
                            text = str(recognition[0])
                            score = safe_float(recognition[1])
                        else:
                            text = str(recognition)
                            score = 0.0

                        blocks.append({
                            "id": len(blocks),
                            "text": text,
                            "bbox_poly": polygon,
                            "bbox": bbox_from_polygon(polygon),
                            "ocr_confidence": round(score, 4),
                        })
                    except Exception:
                        continue

        output_path.write_text(json.dumps(blocks, ensure_ascii=False), encoding="utf-8")
        return 0

    except Exception as exc:
        error = {"error": str(exc), "type": type(exc).__name__}
        output_path.write_text(json.dumps(error, ensure_ascii=False), encoding="utf-8")
        return 1


def run_ocr_with_timeout(image: np.ndarray, timeout_seconds: int = 120) -> Tuple[List[Dict[str, Any]], Optional[str]]:
    temp_dir = Path(tempfile.mkdtemp(prefix="lm_ocr_"))
    image_path = temp_dir / "pdp.jpg"
    result_path = temp_dir / "ocr.json"

    try:
        save_image(image_path, image)
        command = [
            sys.executable,
            str(Path(__file__).resolve()),
            "--ocr-worker",
            str(image_path),
            str(result_path),
        ]

        log(f"[OCR] Starting isolated PaddleOCR worker (timeout: {timeout_seconds}s)...")

        try:
            completed = subprocess.run(
                command,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                text=True,
                timeout=timeout_seconds,
                cwd=str(Path(__file__).resolve().parent),
            )
        except subprocess.TimeoutExpired:
            return [], f"PaddleOCR exceeded {timeout_seconds}s and was terminated."

        if completed.returncode != 0:
            return [], "PaddleOCR worker process encountered an error."

        if not result_path.exists():
            return [], "PaddleOCR produced no output file."

        try:
            data = json.loads(result_path.read_text(encoding="utf-8"))
        except Exception as exc:
            return [], f"Could not parse OCR output: {exc}"

        if isinstance(data, dict) and "error" in data:
            return [], f"PaddleOCR error: {data.get('error')}"

        if not isinstance(data, list):
            return [], "Unexpected OCR data type."

        return data, None

    finally:
        try:
            for child in temp_dir.iterdir():
                child.unlink(missing_ok=True)
            temp_dir.rmdir()
        except Exception:
            pass


# ================================================================
# FORMATTER
# ================================================================

class PipelineFormatter:

    @staticmethod
    def calculate_letter_height_mm(bbox_height_px: float, pdp_height_px: int, physical_height_mm: float) -> float:
        if pdp_height_px <= 0 or physical_height_mm <= 0:
            return 0.0
        return round(max(0.0, bbox_height_px) * physical_height_mm / float(pdp_height_px), 2)

    @staticmethod
    def normalize_text(value: Any) -> str:
        return re.sub(r"[^a-z0-9]", "", str(value).lower())

    def find_ocr_match(self, target: str, ocr_blocks: List[Dict[str, Any]]) -> Tuple[Dict[str, int], float, float]:
        clean_target = self.normalize_text(target)

        if not clean_target:
            return {"x": 0, "y": 0, "w": 0, "h": 0}, 0.0, 0.0

        matching_blocks = []
        target_words = [
            word for word in re.split(r"\W+", str(target).lower()) if len(word) >= 2
        ]

        for block in ocr_blocks:
            clean_ocr = self.normalize_text(block.get("text", ""))
            if not clean_ocr:
                continue

            if clean_ocr in clean_target or clean_target in clean_ocr:
                matching_blocks.append(block)
            elif any(word in clean_ocr for word in target_words):
                matching_blocks.append(block)

        if not matching_blocks:
            return {"x": 0, "y": 0, "w": 0, "h": 0}, 0.0, 0.0

        min_x = min(int(block["bbox"]["x"]) for block in matching_blocks)
        min_y = min(int(block["bbox"]["y"]) for block in matching_blocks)
        max_x = max(
            int(block["bbox"]["x"]) + int(block["bbox"]["w"])
            for block in matching_blocks
        )
        max_y = max(
            int(block["bbox"]["y"]) + int(block["bbox"]["h"])
            for block in matching_blocks
        )

        combined_bbox = {
            "x": max(0, min_x),
            "y": max(0, min_y),
            "w": max(1, max_x - min_x),
            "h": max(1, max_y - min_y),
        }
        avg_confidence = sum(
            safe_float(block.get("ocr_confidence")) for block in matching_blocks
        ) / len(matching_blocks)

        return combined_bbox, 1.0, round(avg_confidence, 4)

    def build_fields(
        self,
        tags: Dict[str, Any],
        ocr_blocks: List[Dict[str, Any]],
        pdp_height_px: int,
        pdp_width_px: int,
        physical_height_mm: float,
    ) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]]]:

        extracted = []
        extra = []

        for field_key, field_val in tags.items():
            if not isinstance(field_val, dict):
                continue

            value_text = str(field_val.get("text") or "").strip()
            if not value_text or value_text.lower() == "null":
                continue

            box_2d = field_val.get("box_2d")
            bbox = {"x": 0, "y": 0, "w": 0, "h": 0}

            if isinstance(box_2d, (list, tuple)) and len(box_2d) == 4:
                try:
                    ymin, xmin, ymax, xmax = box_2d
                    x1 = int(round((float(xmin) / 1000.0) * pdp_width_px))
                    y1 = int(round((float(ymin) / 1000.0) * pdp_height_px))
                    x2 = int(round((float(xmax) / 1000.0) * pdp_width_px))
                    y2 = int(round((float(ymax) / 1000.0) * pdp_height_px))

                    bbox = {
                        "x": max(0, x1),
                        "y": max(0, y1),
                        "w": max(1, x2 - x1),
                        "h": max(1, y2 - y1),
                    }
                except (TypeError, ValueError):
                    bbox = {"x": 0, "y": 0, "w": 0, "h": 0}

            height_mm = self.calculate_letter_height_mm(bbox["h"], pdp_height_px, physical_height_mm)

            extracted.append({
                "field_type": TAG_TO_FIELD_TYPE_MAP.get(field_key, field_key.lower()),
                "text": value_text,
                "confidence": 0.95,
                "bbox": bbox,
                "font_height_mm": height_mm,
                "panel_location": "principal_display_panel",
            })

        return extracted, extra

    def text_measurements(
        self,
        ocr_blocks: List[Dict[str, Any]],
        pdp_height_px: int,
        physical_height_mm: float,
    ) -> List[Dict[str, Any]]:

        result = []
        for block in ocr_blocks:
            height_mm = self.calculate_letter_height_mm(block["bbox"]["h"], pdp_height_px, physical_height_mm)
            result.append({
                "text": block["text"],
                "confidence": block["ocr_confidence"],
                "height_mm": height_mm,
                "min_required_mm": 1.0,
                "is_rule7_compliant": height_mm >= 1.0,
            })
        return result


# ================================================================
# MAIN PIPELINE
# ================================================================

def process_label_verification_multi(
    image_bytes_list: List[bytes],
    api_key: Optional[str] = None,
    product_id: str = "SKU-001",
    category: str = "other",
    sale_channel: str = "retail",
    net_weight_kg: float = 0.273,
    is_imported: bool = False,
    package_surface_type: str = "printed",
    physical_height_mm: float = 260.0,
    physical_width_mm: float = 130.0,
    ocr_timeout_seconds: int = 120,
) -> Dict[str, Any]:

    if not image_bytes_list:
        raise ValueError("No input images were supplied.")

    images = []
    for data in image_bytes_list:
        if not data:
            continue
        array = np.frombuffer(data, dtype=np.uint8)
        image = cv2.imdecode(array, cv2.IMREAD_COLOR)
        if image is not None:
            images.append(image)

    if not images:
        raise ValueError("No valid images could be decoded.")

    log(f"[INFO] Merging {len(images)} packaging photos...")
    combined = stitch_images(images)

    preprocessor = Stage2Preprocessor()
    fallback_package_shape = preprocessor.analyze_package_shape(combined)
    # Resolve API key (use passed value, else fallback to environment/.env)
    final_api_key = api_key or os.getenv("GEMINI_API_KEY")
    if not final_api_key:
        raise ValueError(
            "GEMINI_API_KEY is missing. Please set it in your .env file or pass it directly."
        )

    log("[1/3] Extracting 44 Legal Metrology fields with Gemini...")
    gemini = GeminiExtractor(final_api_key)
    extracted_tags, gemini_package_shape = gemini.extract_tags_and_package_shape(combined)
    package_shape = gemini_package_shape or fallback_package_shape
    pdp, panel_meta = preprocessor.extract_pdp(combined, package_shape)

    log(f"[INFO] Package shape detected: {package_shape}")
    log(f"[INFO] Combined PDP canvas resolution: {pdp.shape[1]}x{pdp.shape[0]} px")

    log("[2/3] Extracting bounding boxes via PaddleOCR...")
    ocr_blocks, ocr_warning = run_ocr_with_timeout(pdp, timeout_seconds=ocr_timeout_seconds)

    if ocr_warning:
        log("[WARNING] " + ocr_warning)

    total_surface_area, required_pdp_area = preprocessor.calculate_required_pdp_area(
        package_shape, physical_height_mm, physical_width_mm
    )

    actual_label_area = max(0.0, physical_height_mm) * max(0.0, physical_width_mm) * panel_meta["pdp_area_ratio"]
    is_small_package = total_surface_area <= 1000.0
    coverage = round(actual_label_area / total_surface_area * 100.0, 2) if total_surface_area > 0 else 0.0

    panel_analysis = {
        "tag": "PRINCIPAL_DISPLAY_PANEL",
        "actual_label_area_sq_mm": round(actual_label_area, 2),
        "required_pdp_area_sq_mm": required_pdp_area,
        "coverage_percentage": coverage,
        "is_small_package_exemption": is_small_package,
        "COMPLIANT": bool(actual_label_area >= required_pdp_area or is_small_package),
    }

    log("[3/3] Bounding box association & metrics rendering...")
    formatter = PipelineFormatter()

    extracted_fields, extra_fields = formatter.build_fields(
        extracted_tags,
        ocr_blocks,
        panel_meta["pdp_height_px"],
        panel_meta["pdp_width_px"],
        physical_height_mm,
    )

    text_measurements = formatter.text_measurements(
        ocr_blocks, panel_meta["pdp_height_px"], physical_height_mm
    )

    # ------------------------------------------------------------
    # ANNOTATION & SEPARATE IMAGE STORAGE
    # ------------------------------------------------------------
    annotated_layers = draw_mandatory_field_boxes(pdp, extracted_fields)

    output_dir = Path(__file__).resolve().parent / "processed_images"
    output_dir.mkdir(parents=True, exist_ok=True)

    saved_paths = []
    for idx, layer_img in enumerate(annotated_layers, start=1):
        suffix = "" if len(annotated_layers) == 1 else f"_layer_{idx}"
        local_file = output_dir / f"{product_id}_stitched_label{suffix}.jpg"
        save_image(local_file, layer_img)
        saved_paths.append(str(local_file.resolve()))

    result_json = {
        "product_id": product_id,
        "category": category,
        "sale_channel": sale_channel,
        "net_weight_kg": net_weight_kg,
        "is_imported": is_imported,
        "package_shape": package_shape,
        "package_surface_type": package_surface_type,
        "local_image_path": saved_paths[0],
        "all_annotated_image_paths": saved_paths,
        "scale_reference_mm": {
            "height_mm": physical_height_mm,
            "width_mm": physical_width_mm,
        },
        "panel_metadata": panel_meta,
        "panel_analysis": panel_analysis,
        "extracted_tags": extracted_tags,
        "extracted_fields": extracted_fields,
    }

    return result_json


# ================================================================
# MULTI-IMAGE DIRECTORY LOADER
# ================================================================

def load_input_images() -> List[bytes]:
    base = Path(__file__).resolve().parent
    valid_extensions = {".jpg", ".jpeg", ".png", ".webp", ".bmp"}

    image_paths = sorted([
        path for path in base.iterdir()
        if path.is_file()
        and path.suffix.lower() in valid_extensions
        and not path.name.startswith("annotated_")
    ])

    image_bytes = []
    for path in image_paths:
        try:
            data = path.read_bytes()
            if data:
                image_bytes.append(data)
                log(f"[INPUT] Loaded packaging photo: {path.name}")
        except Exception as exc:
            log(f"[WARNING] Could not read {path.name}: {exc}")

    return image_bytes


def save_result(result: Dict[str, Any], path: Path) -> None:
    path.write_text(json.dumps(result, indent=2, ensure_ascii=False), encoding="utf-8")


def worker_main() -> int:
    if len(sys.argv) != 4:
        return 2
    return run_paddleocr_worker(Path(sys.argv[2]), Path(sys.argv[3]))


# ================================================================
# MAIN ENTRY
# ================================================================

def main() -> None:
    if len(sys.argv) >= 2 and sys.argv[1] == "--ocr-worker":
        sys.exit(worker_main())

    log("")
    log("=" * 65)
    log("LEGAL METROLOGY MULTI-IMAGE VERIFICATION")
    log("=" * 65)

    if not os.getenv("GEMINI_API_KEY"):
        log("\n[ERROR] GEMINI_API_KEY environment variable is missing.\n")
        log('Run in PowerShell: $env:GEMINI_API_KEY="YOUR_API_KEY"\n')
        return

    image_bytes = load_input_images()

    if not image_bytes:
        log("\n[ERROR] No packaging images found in the directory.")
        log("Add your product photos (e.g. front.jpg, back.jpg, side.jpg) to this folder.\n")
        return

    try:
        result_json = process_label_verification_multi(
            image_bytes_list=image_bytes,
            product_id="SKU-001",
        )

        json_output_path = Path(__file__).resolve().parent / "label_verification_result.json"
        with open(json_output_path, "w", encoding="utf-8") as output_file:
            json.dump(result_json, output_file, indent=2)

        log(f"[INFO] JSON saved successfully to: {json_output_path}")

        log("")
        log("=" * 65)
        log("LABEL VERIFICATION COMPLETE")
        log("=" * 65)
        log(f"Images Processed : {len(image_bytes)}")
        log(f"Package Shape    : {result_json['package_shape']}")
        log(f"PDP Compliant    : {result_json['panel_analysis']['COMPLIANT']}")
        log(f"JSON Result      : {json_output_path.name}")
        log(f"Annotated Output : {result_json['local_image_path']}")
        log("=" * 65)
        log("")

    except Exception as exc:
        log(f"\n[ERROR] Pipeline stopped: {exc}\n")


if __name__ == "__main__":
    main()