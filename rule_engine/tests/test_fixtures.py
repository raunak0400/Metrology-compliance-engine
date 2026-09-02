"""
tests/test_fixtures.py

~20 synthetic ValidationRequest JSON dicts covering edge cases for the
Legal Metrology Rule Engine. These are the canonical fixtures used across
all unit and integration tests.

Each fixture is annotated with what it tests and what the expected outcome is.
"""

from __future__ import annotations

# ---------------------------------------------------------------------------
# Helper: bounding box fixture (not semantically meaningful — placeholder coords)
# ---------------------------------------------------------------------------

def _bbox(x: float = 10.0, y: float = 10.0, w: float = 100.0, h: float = 20.0) -> dict:
    return {"x": x, "y": y, "w": w, "h": h}


def _field(
    field_type: str,
    text: str,
    confidence: float = 0.95,
    font_height_mm: float | None = 1.5,
    panel: str = "principal_display_panel",
) -> dict:
    return {
        "field_type": field_type,
        "text": text,
        "confidence": confidence,
        "bbox": _bbox(),
        "font_height_mm": font_height_mm,
        "panel_location": panel,
    }


# ---------------------------------------------------------------------------
# FIXTURE 1 — Fully compliant retail product (domestic)
# Expected: overall_status = "compliant", violations = []
# ---------------------------------------------------------------------------
FIXTURE_COMPLIANT_RETAIL = {
    "product_id": "TEST-001",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.5,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Acme Industries Pvt. Ltd., 123 Industrial Area, Mumbai - 400001"),
        _field("generic_name", "Refined Sunflower Oil"),
        _field("net_quantity", "500 ml"),
        _field("mrp", "MRP Rs. 120/- (Incl. of all taxes)"),
        _field("mfg_date", "Jan 2024"),
        _field("consumer_care", "Consumer Care: 1800-123-4567 | care@acme.in"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 2 — Missing MRP field entirely
# Expected: violation for MRP_001 / Rule 6(1)(e)
# ---------------------------------------------------------------------------
FIXTURE_MISSING_MRP = {
    "product_id": "TEST-002",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.5,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Acme Industries Pvt. Ltd., 123 Industrial Area, Mumbai"),
        _field("generic_name", "Refined Sunflower Oil"),
        _field("net_quantity", "500 ml"),
        _field("mfg_date", "Jan 2024"),
        _field("consumer_care", "1800-123-4567"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 3 — MRP present but missing "incl. of all taxes" phrase
# Expected: violation for MRP_001 / Rule 6(1)(e) — tax phrase missing
# ---------------------------------------------------------------------------
FIXTURE_MRP_MISSING_TAX_PHRASE = {
    "product_id": "TEST-003",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.3,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Beta Foods Ltd., 45 Commerce St, Delhi - 110001"),
        _field("generic_name", "Biscuits"),
        _field("net_quantity", "200 g"),
        _field("mrp", "MRP Rs. 30/-"),  # Missing "incl. of all taxes"
        _field("mfg_date", "Mar 2024"),
        _field("consumer_care", "9876543210"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 4 — Undersized font on net_quantity field (printed, 0.8mm < 1.0mm)
# Expected: violation for FONT_SIZE_001 / Rule 7
# ---------------------------------------------------------------------------
FIXTURE_FONT_FAIL_PRINTED = {
    "product_id": "TEST-004",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.2,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Gamma Corp, 78 Park Lane, Pune - 411001"),
        _field("generic_name", "Wheat Flour"),
        _field("net_quantity", "200 g", font_height_mm=0.8),  # FAIL: 0.8 < 1.0mm
        _field("mrp", "MRP Rs. 25/- (Incl. of all taxes)"),
        _field("mfg_date", "Feb 2024"),
        _field("consumer_care", "080-12345678"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 5 — Font at 1.5mm on printed packaging (should PASS — threshold is 1.0mm)
# Expected: no font violation
# ---------------------------------------------------------------------------
FIXTURE_FONT_PASS_PRINTED = {
    "product_id": "TEST-005",
    "category": "other",
    "package_shape": "cylindrical",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 1.0,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Delta Beverages, 90 Brew Street, Bengaluru", font_height_mm=1.5),
        _field("generic_name", "Sparkling Water", font_height_mm=1.5),
        _field("net_quantity", "1 l", font_height_mm=1.5),
        _field("mrp", "MRP Rs. 50/- (Incl. of all taxes)", font_height_mm=1.5),
        _field("mfg_date", "Apr 2024", font_height_mm=1.5),
        _field("consumer_care", "care@delta.com", font_height_mm=1.5),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 6 — Missing consumer care field
# Expected: violation for CONSUMER_CARE_001 / Rule 6(1)(f)
# ---------------------------------------------------------------------------
FIXTURE_MISSING_CONSUMER_CARE = {
    "product_id": "TEST-006",
    "category": "cosmetics",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.1,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "GlowUp Cosmetics, 12 Beauty Lane, Chennai - 600001"),
        _field("generic_name", "Face Cream"),
        _field("net_quantity", "50 g"),
        _field("mrp", "MRP Rs. 199/- (Incl. of all taxes)"),
        _field("mfg_date", "Jun 2024"),
        # consumer_care intentionally omitted
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 7 — E-commerce listing (mfg_date should NOT be flagged)
# Expected: overall_status = "compliant", no mfg_date violation
# ---------------------------------------------------------------------------
FIXTURE_ECOMMERCE_NO_MFG_DATE = {
    "product_id": "TEST-007",
    "category": "electronics",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "ecommerce",
    "net_weight_kg": 0.3,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "TechZone Electronics, 55 Circuit Rd, Hyderabad - 500001"),
        _field("generic_name", "USB-C Charging Cable"),
        _field("net_quantity", "1 pcs"),
        _field("mrp", "MRP Rs. 499/- (Incl. of all taxes)"),
        _field("consumer_care", "support@techzone.in"),
        # mfg_date intentionally omitted — exempt for ecommerce
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 8 — Imported product, missing country_of_origin
# Expected: violation for COUNTRY_ORIGIN_001 / Rule 6(1)(b) proviso
# ---------------------------------------------------------------------------
FIXTURE_IMPORTED_MISSING_COUNTRY = {
    "product_id": "TEST-008",
    "category": "electronics",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.5,
    "is_imported": True,
    "extracted_fields": [
        _field("manufacturer", "Import Co. India Pvt Ltd (Importer), 33 Trade Zone, Mumbai"),
        _field("generic_name", "LED Bulb"),
        _field("net_quantity", "1 pcs"),
        _field("mrp", "MRP Rs. 149/- (Incl. of all taxes)"),
        _field("mfg_date", "May 2024"),
        _field("consumer_care", "1800-000-9999"),
        # country_of_origin intentionally omitted
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 9 — Domestic product — country_of_origin correctly not required
# Expected: country_of_origin rule skipped, overall_status = "compliant"
# ---------------------------------------------------------------------------
FIXTURE_DOMESTIC_NO_COUNTRY_REQUIRED = {
    "product_id": "TEST-009",
    "category": "food",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.25,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "NatureFresh Foods, Village Road, Anand, Gujarat"),
        _field("generic_name", "Whole Wheat Flour"),
        _field("net_quantity", "250 g"),
        _field("mrp", "MRP Rs. 20/- (Incl. of all taxes)"),
        _field("mfg_date", "Jul 2024"),
        _field("consumer_care", "info@naturefresh.com"),
        # No country_of_origin field — correct for domestic product
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 10 — Exempt bulk package (>25 kg domestic, category=other)
# Expected: overall_status = "exempt", no violations run
# ---------------------------------------------------------------------------
FIXTURE_EXEMPT_BULK = {
    "product_id": "TEST-010",
    "category": "other",
    "package_shape": "irregular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 30.0,  # > 25 kg threshold
    "is_imported": False,
    "extracted_fields": [],  # No fields — exempt, so not checked
}

# ---------------------------------------------------------------------------
# FIXTURE 11 — Cement (exempt threshold = 50 kg); weight = 40 kg (NOT exempt)
# Expected: rules run normally
# ---------------------------------------------------------------------------
FIXTURE_CEMENT_40KG_NOT_EXEMPT = {
    "product_id": "TEST-011",
    "category": "cement",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 40.0,  # < 50 kg cement threshold — NOT exempt
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "OPC Cement Ltd., Industrial Estate, Rajasthan"),
        _field("generic_name", "Ordinary Portland Cement"),
        _field("net_quantity", "40 kg"),
        _field("mrp", "MRP Rs. 380/- (Incl. of all taxes)"),
        _field("mfg_date", "Aug 2024"),
        _field("consumer_care", "1800-456-7890"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 12 — Food category product (mfg_date and net_quantity advisory)
# Expected: violations with severity="advisory" if those fail; passes if present
# ---------------------------------------------------------------------------
FIXTURE_FOOD_ADVISORY = {
    "product_id": "TEST-012",
    "category": "food",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.5,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Heritage Dairy, Milk Colony, Amul Road, Anand"),
        _field("generic_name", "Cow Milk"),
        _field("net_quantity", "500 ml"),
        _field("mrp", "MRP Rs. 28/- (Incl. of all taxes)"),
        _field("mfg_date", "Sep 2024"),
        _field("consumer_care", "consumer@heritage.com"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 13 — Low confidence field (lands in manual_review, NOT violations)
# consumer_care field has confidence 0.50 < 0.75 threshold
# Expected: consumer_care in manual_review_required, NOT in violations
# ---------------------------------------------------------------------------
FIXTURE_LOW_CONFIDENCE_CONSUMER_CARE = {
    "product_id": "TEST-013",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.3,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Sigma Products, 200 Warehouse Road, Noida"),
        _field("generic_name", "Plastic Storage Bin"),
        _field("net_quantity", "1 pcs"),
        _field("mrp", "MRP Rs. 99/- (Incl. of all taxes)"),
        _field("mfg_date", "Oct 2024"),
        _field("consumer_care", "9876543210", confidence=0.50),  # LOW CONFIDENCE
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 14 — Field wrongly placed outside principal display panel
# net_quantity placed on "other" panel
# Expected: placement violation for PLACEMENT_001 / Rule 9
# ---------------------------------------------------------------------------
FIXTURE_WRONG_PANEL_PLACEMENT = {
    "product_id": "TEST-014",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.2,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Omega Packaging, 15 Factory Lane, Surat"),
        _field("generic_name", "Bath Soap"),
        _field("net_quantity", "150 g", panel="other"),  # WRONG PANEL
        _field("mrp", "MRP Rs. 45/- (Incl. of all taxes)"),
        _field("mfg_date", "Nov 2024"),
        _field("consumer_care", "care@omega.com"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 15 — Embossed packaging, font 1.5mm (FAIL — threshold is 2.0mm)
# Expected: font violation for FONT_SIZE_001
# ---------------------------------------------------------------------------
FIXTURE_FONT_FAIL_EMBOSSED = {
    "product_id": "TEST-015",
    "category": "other",
    "package_shape": "cylindrical",
    "package_surface_type": "embossed_molded_blown",
    "sale_channel": "retail",
    "net_weight_kg": 0.75,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "ClearGlass Pvt Ltd, Firozabad, Uttar Pradesh", font_height_mm=2.5),
        _field("generic_name", "Premium Mineral Water", font_height_mm=2.5),
        _field("net_quantity", "750 ml", font_height_mm=1.5),  # FAIL: 1.5 < 2.0mm for embossed
        _field("mrp", "MRP Rs. 30/- (Incl. of all taxes)", font_height_mm=2.5),
        _field("mfg_date", "Dec 2024", font_height_mm=2.5),
        _field("consumer_care", "1800-111-2222", font_height_mm=2.5),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 16 — Printed packaging, font 1.5mm (PASS — threshold is 1.0mm)
# Expected: no font violation
# ---------------------------------------------------------------------------
FIXTURE_FONT_PASS_PRINTED_1_5 = {
    "product_id": "TEST-016",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.4,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "PrintPack Ltd, 77 Paper Mill Rd, Kolkata", font_height_mm=1.5),
        _field("generic_name", "Paper Napkins", font_height_mm=1.5),
        _field("net_quantity", "100 pcs", font_height_mm=1.5),
        _field("mrp", "MRP Rs. 60/- (Incl. of all taxes)", font_height_mm=1.5),
        _field("mfg_date", "Jan 2025", font_height_mm=1.5),
        _field("consumer_care", "support@printpack.com", font_height_mm=1.5),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 17 — Multiple violations: missing MRP + missing consumer care
# Expected: 2 violations — MRP_001 and CONSUMER_CARE_001
# ---------------------------------------------------------------------------
FIXTURE_MULTIPLE_VIOLATIONS = {
    "product_id": "TEST-017",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.1,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "No-Label Products, Back Lane, Unknown City"),
        _field("generic_name", "Some Product"),
        _field("net_quantity", "100 g"),
        _field("mfg_date", "Feb 2025"),
        # mrp and consumer_care intentionally omitted
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 18 — Imported product WITH country_of_origin (compliant)
# Expected: overall_status = "compliant"
# ---------------------------------------------------------------------------
FIXTURE_IMPORTED_WITH_COUNTRY = {
    "product_id": "TEST-018",
    "category": "electronics",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.3,
    "is_imported": True,
    "extracted_fields": [
        _field("manufacturer", "TechImports India Pvt Ltd (Importer), 101 Import Hub, Mumbai"),
        _field("generic_name", "Wireless Earbuds"),
        _field("net_quantity", "1 pcs"),
        _field("mrp", "MRP Rs. 1299/- (Incl. of all taxes)"),
        _field("mfg_date", "Mar 2025"),
        _field("consumer_care", "1800-999-0000"),
        _field("country_of_origin", "Made in China"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 19 — Net quantity with invalid/non-standard unit ("handful")
# Expected: violation for NET_QUANTITY_001 / Rule 6(1)(c) & Rule 11
# ---------------------------------------------------------------------------
FIXTURE_INVALID_QUANTITY_UNIT = {
    "product_id": "TEST-019",
    "category": "food",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.05,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "TraditionalFoods Co., Rural Lane, Mysore"),
        _field("generic_name", "Mixed Spices"),
        _field("net_quantity", "2 handfuls"),  # INVALID unit
        _field("mrp", "MRP Rs. 15/- (Incl. of all taxes)"),
        _field("mfg_date", "Apr 2025"),
        _field("consumer_care", "tradifood@gmail.com"),
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 20 — Consumer care field present but no phone/email (just text)
# Expected: violation for CONSUMER_CARE_001 — no valid contact info
# ---------------------------------------------------------------------------
FIXTURE_CONSUMER_CARE_NO_CONTACT = {
    "product_id": "TEST-020",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.5,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Vague Corp, Some Place, India"),
        _field("generic_name", "Mystery Product"),
        _field("net_quantity", "500 g"),
        _field("mrp", "MRP Rs. 100/- (Incl. of all taxes)"),
        _field("mfg_date", "May 2025"),
        _field("consumer_care", "For queries, please contact us"),  # No phone/email
    ],
}

# ---------------------------------------------------------------------------
# FIXTURE 21 — Exempt bulk cement package (>50 kg)
# Expected: overall_status = "exempt"
# ---------------------------------------------------------------------------
FIXTURE_EXEMPT_CEMENT_BULK = {
    "product_id": "TEST-021",
    "category": "cement",
    "package_shape": "irregular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 55.0,  # > 50 kg cement threshold
    "is_imported": False,
    "extracted_fields": [],
}

# ---------------------------------------------------------------------------
# FIXTURE 22 — Exempt bulk package via VOLUME (>25 L, 'other' category)
# A 30-litre oil drum — exempt by volume, not weight.
# Expected: overall_status = "exempt"
# ---------------------------------------------------------------------------
FIXTURE_EXEMPT_VOLUME = {
    "product_id": "TEST-022",
    "category": "other",
    "package_shape": "cylindrical",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": None,       # weight not provided — only volume drives exemption
    "net_volume_l": 30.0,        # > 25 L threshold — EXEMPT
    "is_imported": False,
    "extracted_fields": [],
}

# ---------------------------------------------------------------------------
# FIXTURE 23 — Missing generic_name field
# Expected: violation for GENERIC_NAME_001 / Rule 6(1)(a) & Rule 2(l)
# ---------------------------------------------------------------------------
FIXTURE_MISSING_GENERIC_NAME = {
    "product_id": "TEST-023",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.3,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "NoName Corp, 1 Unknown Lane, Delhi"),
        # generic_name intentionally omitted
        _field("net_quantity", "300 g"),
        _field("mrp", "MRP Rs. 50/- (Incl. of all taxes)"),
        _field("mfg_date", "Jun 2025"),
        _field("consumer_care", "help@noname.in"),
    ],
}

# ---------------------------------------------------------------------------
# All fixtures as an indexed dict for easy parametrization in pytest
# ---------------------------------------------------------------------------
ALL_FIXTURES = {
    "compliant_retail": FIXTURE_COMPLIANT_RETAIL,
    "missing_mrp": FIXTURE_MISSING_MRP,
    "mrp_missing_tax_phrase": FIXTURE_MRP_MISSING_TAX_PHRASE,
    "font_fail_printed": FIXTURE_FONT_FAIL_PRINTED,
    "font_pass_printed": FIXTURE_FONT_PASS_PRINTED,
    "missing_consumer_care": FIXTURE_MISSING_CONSUMER_CARE,
    "ecommerce_no_mfg_date": FIXTURE_ECOMMERCE_NO_MFG_DATE,
    "imported_missing_country": FIXTURE_IMPORTED_MISSING_COUNTRY,
    "domestic_no_country_required": FIXTURE_DOMESTIC_NO_COUNTRY_REQUIRED,
    "exempt_bulk": FIXTURE_EXEMPT_BULK,
    "cement_40kg_not_exempt": FIXTURE_CEMENT_40KG_NOT_EXEMPT,
    "food_advisory": FIXTURE_FOOD_ADVISORY,
    "low_confidence_consumer_care": FIXTURE_LOW_CONFIDENCE_CONSUMER_CARE,
    "wrong_panel_placement": FIXTURE_WRONG_PANEL_PLACEMENT,
    "font_fail_embossed": FIXTURE_FONT_FAIL_EMBOSSED,
    "font_pass_printed_1_5": FIXTURE_FONT_PASS_PRINTED_1_5,
    "multiple_violations": FIXTURE_MULTIPLE_VIOLATIONS,
    "imported_with_country": FIXTURE_IMPORTED_WITH_COUNTRY,
    "invalid_quantity_unit": FIXTURE_INVALID_QUANTITY_UNIT,
    "consumer_care_no_contact": FIXTURE_CONSUMER_CARE_NO_CONTACT,
    "exempt_cement_bulk": FIXTURE_EXEMPT_CEMENT_BULK,
    "exempt_volume": FIXTURE_EXEMPT_VOLUME,
    "missing_generic_name": FIXTURE_MISSING_GENERIC_NAME,
}

# ---------------------------------------------------------------------------
# TAG BRIDGE FIXTURES (FIXTURE 24–27)
# These test the extracted_tags → ExtractedField synthesis path.
# ---------------------------------------------------------------------------

# FIXTURE 24 — Tags only, no extracted_fields
# The engine must synthesize all 7 fields from extracted_tags and produce
# a compliant result (Parachute Coconut Oil scenario, domestic).
FIXTURE_TAGS_ONLY_DOMESTIC = {
    "product_id": "TAG-BRIDGE-001",
    "category": "other",
    "package_shape": "cylindrical",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.273,
    "is_imported": False,
    "extracted_fields": [],          # intentionally empty — bridge must fill
    "extracted_tags": {
        "MANUFACTURER_NAME": "Marico Ltd.",
        "MANUFACTURER_ADDRESS": "7th Floor, Grande Palladium, 175 CST Road, Kalina, Santacruz (East), Mumbai - 400 098, MH.",
        "PACKER_NAME": "Sowparnika Industries Private Limited",
        "PACKER_ADDRESS": "P.V.Palayam, Perundurai, Erode - 638 052, TN",
        "IMPORTER_NAME": None,
        "IMPORTER_ADDRESS": None,
        "COUNTRY_OF_ORIGIN": "India",
        "PRODUCT_COMMON_NAME": "100% Pure Coconut Oil",
        "NET_QUANTITY": "300 ml + 30 ml FREE",
        "NET_QUANTITY_UNIT": "ml",
        "MANUFACTURE_MONTH": "MAY",
        "MANUFACTURE_YEAR": "2026",
        "MRP": "₹222",
        "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
        "CONSUMER_CARE_PHONE": "1800222248",
        "CONSUMER_CARE_EMAIL": "csc@marico.com",
    },
    "text_measurements": [
        {
            "text": "Marico Ltd.",
            "confidence": 0.96,
            "height_mm": 2.1,
            "min_required_mm": 1.0,
            "is_rule7_compliant": True,
        },
        {
            "text": "300 ml",
            "confidence": 0.97,
            "height_mm": 3.15,
            "min_required_mm": 1.0,
            "is_rule7_compliant": True,
        },
    ],
}

# FIXTURE 25 — extracted_fields present AND extracted_tags present
# extracted_fields has manufacturer; tags have the rest.
# Merge result: manufacturer from fields (priority), rest synthesized from tags.
FIXTURE_TAGS_MERGE_WITH_FIELDS = {
    "product_id": "TAG-BRIDGE-002",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.5,
    "is_imported": False,
    "extracted_fields": [
        _field("manufacturer", "Acme Industries Pvt. Ltd., 123 Road, Mumbai - 400001"),
        _field("generic_name", "Refined Sunflower Oil"),
        # mfg_date, mrp, consumer_care, net_quantity come from tags
    ],
    "extracted_tags": {
        "MANUFACTURER_NAME": "Should Be Ignored Because extracted_fields Has It",
        "NET_QUANTITY": "500 ml",
        "MANUFACTURE_MONTH": "Jan",
        "MANUFACTURE_YEAR": "2025",
        "MRP": "Rs. 120",
        "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
        "CONSUMER_CARE_PHONE": "9876543210",
    },
}

# FIXTURE 26 — Imported product with IMPORTER_NAME/ADDRESS in tags (not null)
# Engine must synthesize an 'importer' field, manufacturer rule must pass.
FIXTURE_TAGS_IMPORTED_WITH_IMPORTER = {
    "product_id": "TAG-BRIDGE-003",
    "category": "electronics",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.3,
    "is_imported": True,
    "extracted_fields": [],
    "extracted_tags": {
        "MANUFACTURER_NAME": "SomeOEM Corp.",
        "MANUFACTURER_ADDRESS": "Shanghai Industrial Zone, China",
        "IMPORTER_NAME": "TechImports India Pvt. Ltd.",
        "IMPORTER_ADDRESS": "101 Import Hub, Andheri East, Mumbai - 400069",
        "PRODUCT_COMMON_NAME": "USB-C Charging Cable",
        "NET_QUANTITY": "1 pcs",
        "MANUFACTURE_MONTH": "March",
        "MANUFACTURE_YEAR": "2025",
        "MRP": "MRP Rs. 499",
        "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
        "CONSUMER_CARE_PHONE": "1800-999-1111",
        "COUNTRY_OF_ORIGIN": "China",
    },
}

# FIXTURE 27 — Multi-pack notation via tags ("2 x 325 ml")
# Tests that the bridge + net_quantity rule correctly handle multi-pack text.
FIXTURE_TAGS_MULTIPACK = {
    "product_id": "TAG-BRIDGE-004",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.65,
    "is_imported": False,
    "extracted_fields": [],
    "extracted_tags": {
        "MANUFACTURER_NAME": "HairCare Ltd.",
        "MANUFACTURER_ADDRESS": "12 Beauty Road, Pune - 411001",
        "PRODUCT_COMMON_NAME": "Shampoo",
        "NET_QUANTITY": "2 x 325 ml",
        "MANUFACTURE_MONTH": "FEB",
        "MANUFACTURE_YEAR": "2025",
        "MRP": "₹399",
        "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
        "CONSUMER_CARE_EMAIL": "care@haircare.in",
    },
}

# ---------------------------------------------------------------------------
# API DEMO REGRESSION FIXTURES (FIXTURE 28–32)
# Derived from the 6 real API call observations made during integration testing.
# ---------------------------------------------------------------------------

# FIXTURE 28 — BISCUIT-TAGS-005: Dual-confidence MRP with correct text on the
# lower-confidence extraction.
#
# Design observation (not a bug): confidence-based selection means the rule
# engine uses the highest-confidence mrp extraction. If that extraction is
# missing the tax phrase, MRP_001 fires — even if a lower-confidence extraction
# would have passed. This is the documented tradeoff of confidence-based selection.
# Expected: MRP_001 violation fired (high-conf "MRP ₹40/-" used, sub-threshold
# "₹40 (Incl. of all taxes)" at 0.74 discarded).
FIXTURE_BISCUIT_DUAL_CONFIDENCE_MRP = {
    "product_id": "BISCUIT-TAGS-005",
    "category": "food",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.1,
    "is_imported": False,
    "extracted_fields": [
        # Sub-threshold MRP: has correct tax phrase but confidence too low (0.74 < 0.75)
        _field("mrp", "₹40 (Incl. of all taxes)", confidence=0.74),
        # High-confidence MRP: missing tax phrase — this is what the rule sees
        _field("mrp", "MRP ₹40/-", confidence=0.91),
        _field("manufacturer", "Britannia Industries, Kolkata - 700001"),
        _field("generic_name", "Biscuits"),
        _field("net_quantity", "100 g"),
        _field("mfg_date", "Jan 2025"),
        _field("consumer_care", "1800-103-1813"),
    ],
}

# FIXTURE 29 — SPICE-TAGS-002: Ambiguous text_measurements (two entries both
# matching "100 g" as a substring, but the bridge must pick the closest-length
# match — not naively the first list entry).
#
# text_measurements:
#   [0] "100 g pack date info" at 1.4 mm  — longer string, comes FIRST in list
#   [1] "100 g"               at 2.6 mm  — exact match, comes SECOND
#
# _find_font_height must return 2.6 (exact/closest-length wins).
# Check: "100 g pack date info" in "100 g" → False (longer string never a
# substring of a shorter string). "100 g" in "100 g" → True. So only one match
# exists — the exact match — and 2.6mm is returned.
FIXTURE_SPICE_AMBIGUOUS_FONT_MEASUREMENT = {
    "product_id": "SPICE-TAGS-002",
    "category": "food",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.05,
    "is_imported": False,
    "extracted_fields": [],
    "extracted_tags": {
        "MANUFACTURER_NAME": "SpiceCo Ltd.",
        "MANUFACTURER_ADDRESS": "45 Spice Market, Kochi - 682001",
        "PRODUCT_COMMON_NAME": "Ground Turmeric",
        "NET_QUANTITY": "100 g",
        "MANUFACTURE_MONTH": "Dec",
        "MANUFACTURE_YEAR": "2024",
        "MRP": "₹35",
        "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
        "CONSUMER_CARE_PHONE": "080-22334455",
    },
    "text_measurements": [
        # Deliberately placed FIRST in list — must NOT win over exact match
        {
            "text": "100 g pack date info",
            "confidence": 0.82,
            "height_mm": 1.4,
            "min_required_mm": 1.0,
            "is_rule7_compliant": True,
        },
        # Exact match — must win (closest-length algorithm)
        {
            "text": "100 g",
            "confidence": 0.97,
            "height_mm": 2.6,
            "min_required_mm": 1.0,
            "is_rule7_compliant": True,
        },
    ],
}

# FIXTURE 30 — CANDLE-TAGS-006: No text_measurements overlap with synthesized
# field text — font_height_mm must resolve to None and FONT_SIZE_001 must be
# silently skipped (not defaulting to pass or throwing an error).
FIXTURE_CANDLE_NO_MEASUREMENT_MATCH = {
    "product_id": "CANDLE-TAGS-006",
    "category": "other",
    "package_shape": "cylindrical",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.15,
    "is_imported": False,
    "extracted_fields": [],
    "extracted_tags": {
        "MANUFACTURER_NAME": "WaxWorks Pvt. Ltd.",
        "MANUFACTURER_ADDRESS": "7 Candle Lane, Ahmedabad - 380001",
        "PRODUCT_COMMON_NAME": "Scented Candle",
        "NET_QUANTITY": "150 g",
        "MANUFACTURE_MONTH": "OCT",
        "MANUFACTURE_YEAR": "2024",
        "MRP": "₹120",
        "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
        "CONSUMER_CARE_PHONE": "079-11223344",
    },
    # text_measurements has NO overlap with "150 g" → font_height_mm must be null
    "text_measurements": [
        {
            "text": "wick length 5cm",
            "confidence": 0.90,
            "height_mm": 1.5,
            "min_required_mm": 1.0,
            "is_rule7_compliant": True,
        },
        {
            "text": "burn time 40hrs",
            "confidence": 0.88,
            "height_mm": 1.2,
            "min_required_mm": 1.0,
            "is_rule7_compliant": True,
        },
    ],
}

# FIXTURE 31 — COOKINGOIL-TAGS-001: Imported product with explicit importer
# tags (priority ①), bonus-pack quantity ("1 l + 100 ml FREE"), food-category
# advisory downgrade, and country_of_origin satisfied. Must be compliant.
FIXTURE_COOKINGOIL_IMPORTED_BONUS_PACK = {
    "product_id": "COOKINGOIL-TAGS-001",
    "category": "food",
    "package_shape": "cylindrical",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 1.1,
    "is_imported": True,
    "extracted_fields": [],
    "extracted_tags": {
        "MANUFACTURER_NAME": "OilPress Co.",
        "MANUFACTURER_ADDRESS": "Athens Industrial Zone, Greece",
        "IMPORTER_NAME": "PureOil India Pvt. Ltd.",
        "IMPORTER_ADDRESS": "22 Trade Centre, Mumbai - 400001",
        "PRODUCT_COMMON_NAME": "Extra Virgin Olive Oil",
        "NET_QUANTITY": "1 l + 100 ml FREE",
        "NET_QUANTITY_UNIT": "l",
        "MANUFACTURE_MONTH": "JUN",
        "MANUFACTURE_YEAR": "2025",
        "MRP": "₹549",
        "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
        "CONSUMER_CARE_PHONE": "1800-111-9090",
        "COUNTRY_OF_ORIGIN": "Greece",
    },
    "text_measurements": [
        {
            "text": "1 l + 100 ml FREE",
            "confidence": 0.96,
            "height_mm": 2.8,
            "min_required_mm": 1.0,
            "is_rule7_compliant": True,
        },
    ],
}

# FIXTURE 32 — SHAMPOO-TAGS-004: E-commerce listing missing mfg_date on the
# physical label. mfg_date MUST fire (Rule 6(10) does not exempt physical label).
# Direct regression test for the Rule 6(10) fix.
FIXTURE_SHAMPOO_ECOMMERCE_MISSING_MFGDATE = {
    "product_id": "SHAMPOO-TAGS-004",
    "category": "other",
    "package_shape": "cylindrical",
    "package_surface_type": "printed",
    "sale_channel": "ecommerce",
    "net_weight_kg": 0.325,
    "is_imported": False,
    "extracted_fields": [],
    "extracted_tags": {
        "MANUFACTURER_NAME": "HairCare Ltd.",
        "MANUFACTURER_ADDRESS": "12 Beauty Road, Pune - 411001",
        "PRODUCT_COMMON_NAME": "Anti-Dandruff Shampoo",
        "NET_QUANTITY": "325 ml",
        "MRP": "₹299",
        "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
        "CONSUMER_CARE_PHONE": "1800-200-1234",
        # MANUFACTURE_MONTH and MANUFACTURE_YEAR intentionally omitted
        # to confirm mfg_date violation fires even for ecommerce
    },
}
