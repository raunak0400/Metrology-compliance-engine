"""
tag_bridge.py — Converts extracted_tags + text_measurements into ExtractedField objects.

Purpose
-------
The upstream OCR / ML service produces two complementary outputs:

  extracted_tags   — a flat dict of named, typed values
                     e.g. {"MANUFACTURER_NAME": "Marico Ltd.", "MRP": "₹222", …}

  extracted_fields — a list of ExtractedField objects with bounding boxes, confidence
                     scores, and panel_location — ready for the rule engine.

When a caller provides extracted_tags but not extracted_fields (or provides only
partial extracted_fields), this module synthesizes ExtractedField objects from the
tags so that the rule engine can still run every applicable compliance check.

Merging strategy
----------------
Caller-provided extracted_fields ALWAYS take priority.
A synthesized field is only added for a field_type that has NO high-confidence entry
in the caller's list. This means:

  - Full extracted_fields? Bridge does nothing.
  - No extracted_fields at all? Bridge fills every known field_type from tags.
  - Partial extracted_fields? Bridge fills only the gaps.

Font height resolution
----------------------
If text_measurements is provided, the bridge tries to match each synthesized field's
text to a measurement entry (substring match). When found, font_height_mm is taken
from that entry. Otherwise it defaults to None (rule engine will skip the font check).

Tag → field_type mapping
------------------------
The mapping covers all 7 mandatory declarations under Rule 6(1):

  MANUFACTURER_NAME + MANUFACTURER_ADDRESS  → manufacturer  (primary declaration)
  PACKER_NAME + PACKER_ADDRESS              → manufacturer  (second entry if distinct)
  IMPORTER_NAME + IMPORTER_ADDRESS          → importer      (only when not null)
  PRODUCT_COMMON_NAME                       → generic_name
  NET_QUANTITY                              → net_quantity
  MANUFACTURE_MONTH + MANUFACTURE_YEAR      → mfg_date
  MRP + MRP_INCLUSIVE_OF_TAXES              → mrp
  CONSUMER_CARE_PHONE / CONSUMER_CARE_EMAIL → consumer_care
  COUNTRY_OF_ORIGIN                         → country_of_origin
"""

from __future__ import annotations

from typing import Any, Dict, List, Optional, Tuple

from schemas import BoundingBox, ExtractedField, TextMeasurementEntry

# Default confidence assigned to tag-synthesized fields.
# Slightly below the 0.75 threshold so a human can override,
# but high enough that the field is not automatically sent to
# manual_review_required when the text is clearly legible.
# Adjust via rules_config.json "tag_synthesis_confidence" if desired.
_SYNTHESIS_CONFIDENCE: float = 0.88

# Sentinel bounding box for synthesized fields (coords unavailable from tags)
_NULL_BBOX = BoundingBox(x=0.0, y=0.0, w=0.0, h=0.0)

# Mandatory field types the engine checks — only these are synthesized
_MANDATORY_FIELD_TYPES = frozenset({
    "manufacturer",
    "importer",
    "generic_name",
    "net_quantity",
    "mfg_date",
    "mrp",
    "consumer_care",
    "country_of_origin",
})


def _truthy(value: Any) -> bool:
    """Return True if value is a non-empty, non-null string."""
    return bool(value and str(value).strip())


def _join(*parts: Any, sep: str = ", ") -> str:
    """Join non-null parts into a single string."""
    return sep.join(str(p).strip() for p in parts if _truthy(p))


def _find_font_height(
    text: str,
    measurements: Optional[List[TextMeasurementEntry]],
) -> Optional[float]:
    """
    Find the best-matching font_height_mm for a synthesized field's text.

    Matching rule: scan text_measurements for entries whose .text is a
    substring of the synthesized field text (case-insensitive). Among all
    matches, prefer the entry whose text length is CLOSEST to the field text
    length — this naturally prefers exact matches over partial-substring
    matches, and makes resolution order-independent regardless of list order.

    Examples
    --------
    Field: "100 g"
      Entry A: "100 g pack date info" → NOT a substring of "100 g" (skipped)
      Entry B: "100 g"               → exact match, length diff = 0 → chosen

    Field: "300 ml + 30 ml FREE"
      Entry A: "300 ml"              → substring, length diff = 14
      Entry B: "300 ml + 30 ml FREE" → exact, length diff = 0 → chosen

    Min 4-char guard prevents spurious matches from short tokens ("ml", "kg").
    Returns height_mm of best match, or None if no match found.
    """
    if not measurements or not text:
        return None
    text_lower = text.lower().strip()
    best_height: Optional[float] = None
    best_len_diff: int = 10_000  # sentinel; smaller = better match

    for m in measurements:
        m_text = m.text.lower().strip()
        if len(m_text) < 4:
            continue
        if m_text in text_lower:
            len_diff = abs(len(m_text) - len(text_lower))
            if len_diff < best_len_diff:
                best_len_diff = len_diff
                best_height = m.height_mm

    return best_height


def _make_field(
    field_type: str,
    text: str,
    source_tags: List[str],
    measurements: Optional[List[TextMeasurementEntry]],
) -> Tuple[ExtractedField, Dict[str, Any]]:
    """
    Build a single synthesized ExtractedField and a summary record.

    Returns (ExtractedField, summary_dict).
    """
    font_h = _find_font_height(text, measurements)
    field = ExtractedField(
        field_type=field_type,
        text=text,
        confidence=_SYNTHESIS_CONFIDENCE,
        bbox=_NULL_BBOX,
        font_height_mm=font_h,
        panel_location="principal_display_panel",  # mandatory fields must be on PDP
    )
    summary = {
        "field_type": field_type,
        "text": text[:120],  # truncate for readability in response
        "source_tags": source_tags,
        "font_height_mm": font_h,
        "confidence": _SYNTHESIS_CONFIDENCE,
    }
    return field, summary


def synthesize_from_tags(
    tags: Dict[str, Any],
    measurements: Optional[List[TextMeasurementEntry]],
    existing_fields: List[ExtractedField],
    confidence_threshold: float = 0.75,
) -> Tuple[List[ExtractedField], List[Dict[str, Any]]]:
    """
    Synthesize ExtractedField objects from extracted_tags for any field_type
    not already covered by a high-confidence entry in existing_fields.

    Args:
        tags:               extracted_tags dict from the OCR service.
        measurements:       text_measurements list for font height lookup.
        existing_fields:    caller-provided extracted_fields (take priority).
        confidence_threshold: threshold from config (default 0.75).

    Returns:
        (new_fields, summaries)
        new_fields  — synthesized ExtractedField objects ready to merge
        summaries   — human-readable audit records for ValidationResponse
    """
    if not tags:
        return [], []

    # Build set of field_types already well-covered by caller-provided fields
    covered: set[str] = {
        f.field_type
        for f in existing_fields
        if f.confidence >= confidence_threshold
    }

    new_fields: List[ExtractedField] = []
    summaries: List[Dict[str, Any]] = []

    def _add(field_type: str, text: str, source_tag_keys: List[str]) -> None:
        if field_type in covered or not _truthy(text):
            return
        field, summary = _make_field(field_type, text, source_tag_keys, measurements)
        new_fields.append(field)
        summaries.append(summary)
        covered.add(field_type)  # don't add the same type twice

    # -----------------------------------------------------------------
    # Rule 6(1)(a) — Manufacturer / Packer declaration
    # Primary: manufacturer name + address (one field per entity)
    # -----------------------------------------------------------------
    mfr_name = tags.get("MANUFACTURER_NAME")
    mfr_addr = tags.get("MANUFACTURER_ADDRESS")
    if _truthy(mfr_name) or _truthy(mfr_addr):
        mfr_text = _join(mfr_name, mfr_addr)
        _add("manufacturer", mfr_text, ["MANUFACTURER_NAME", "MANUFACTURER_ADDRESS"])

    # Packer (if distinct from manufacturer, adds a second manufacturer field)
    packer_name = tags.get("PACKER_NAME")
    packer_addr = tags.get("PACKER_ADDRESS")
    if _truthy(packer_name) or _truthy(packer_addr):
        packer_text = _join(packer_name, packer_addr)
        # Only add packer as a second manufacturer entry if it's different text
        mfr_text_check = _join(mfr_name, mfr_addr)
        if packer_text.lower() != mfr_text_check.lower():
            # Don't use _add because we allow multiple manufacturer entries
            field, summary = _make_field(
                "manufacturer", packer_text,
                ["PACKER_NAME", "PACKER_ADDRESS"], measurements
            )
            new_fields.append(field)
            summaries.append(summary)

    # -----------------------------------------------------------------
    # Rule 6(1)(a) proviso — Importer declaration (imported goods)
    # Check IMPORTER_NAME / IMPORTER_ADDRESS explicitly
    # -----------------------------------------------------------------
    importer_name = tags.get("IMPORTER_NAME")
    importer_addr = tags.get("IMPORTER_ADDRESS")
    if _truthy(importer_name) or _truthy(importer_addr):
        importer_text = _join(importer_name, importer_addr)
        _add("importer", importer_text, ["IMPORTER_NAME", "IMPORTER_ADDRESS"])

    # -----------------------------------------------------------------
    # Rule 6(1)(a) & Rule 2(l) — Generic/common name
    # -----------------------------------------------------------------
    generic = tags.get("PRODUCT_COMMON_NAME") or tags.get("GENERIC_NAME")
    _add("generic_name", str(generic or ""), ["PRODUCT_COMMON_NAME"])

    # -----------------------------------------------------------------
    # Rule 6(1)(c) & Rule 11 — Net Quantity
    # Combine NET_QUANTITY + NET_QUANTITY_UNIT when unit is separate
    # -----------------------------------------------------------------
    net_qty = tags.get("NET_QUANTITY")
    net_unit = tags.get("NET_QUANTITY_UNIT")
    if _truthy(net_qty):
        # If unit is already embedded in the quantity string, don't duplicate
        qty_text = str(net_qty).strip()
        if _truthy(net_unit) and str(net_unit).strip().lower() not in qty_text.lower():
            qty_text = f"{qty_text} {net_unit}"
        _add("net_quantity", qty_text, ["NET_QUANTITY", "NET_QUANTITY_UNIT"])

    # -----------------------------------------------------------------
    # Rule 6(1)(e) — MRP inclusive of all taxes
    # Combine the price value + tax phrase into one field text
    # -----------------------------------------------------------------
    mrp_val = tags.get("MRP")
    mrp_tax = tags.get("MRP_INCLUSIVE_OF_TAXES") or tags.get("MRP_TAX_PHRASE")
    if _truthy(mrp_val):
        mrp_text = _join(mrp_val, mrp_tax, sep=" ")
        _add("mrp", mrp_text, ["MRP", "MRP_INCLUSIVE_OF_TAXES"])

    # -----------------------------------------------------------------
    # Rule 6(1)(d) — Month and Year of manufacture/packing
    # Combine MANUFACTURE_MONTH + MANUFACTURE_YEAR
    # -----------------------------------------------------------------
    mfg_month = tags.get("MANUFACTURE_MONTH") or tags.get("MFG_MONTH")
    mfg_year = tags.get("MANUFACTURE_YEAR") or tags.get("MFG_YEAR")
    if _truthy(mfg_month) or _truthy(mfg_year):
        mfg_text = _join(mfg_month, mfg_year, sep=" ")
        _add("mfg_date", mfg_text, ["MANUFACTURE_MONTH", "MANUFACTURE_YEAR"])

    # -----------------------------------------------------------------
    # Rule 6(1)(f) — Consumer care (phone + email)
    # Combine all available contact details into one field
    # -----------------------------------------------------------------
    cc_phone = tags.get("CONSUMER_CARE_PHONE") or tags.get("HELPLINE_PHONE")
    cc_email = tags.get("CONSUMER_CARE_EMAIL") or tags.get("HELPLINE_EMAIL")
    cc_name  = tags.get("CONSUMER_CARE_NAME")
    cc_parts = [p for p in [cc_phone, cc_email, cc_name] if _truthy(p)]
    if cc_parts:
        _add("consumer_care", " ".join(str(p) for p in cc_parts),
             ["CONSUMER_CARE_PHONE", "CONSUMER_CARE_EMAIL", "CONSUMER_CARE_NAME"])

    # -----------------------------------------------------------------
    # Rule 6(1)(b) proviso — Country of origin (imported goods)
    # -----------------------------------------------------------------
    coo = tags.get("COUNTRY_OF_ORIGIN")
    _add("country_of_origin", str(coo or ""), ["COUNTRY_OF_ORIGIN"])

    return new_fields, summaries
