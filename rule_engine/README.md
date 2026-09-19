# Legal Metrology Rule Engine

A standalone **FastAPI microservice** that performs deterministic, rule-based compliance checking for packaged commodities under India's **Legal Metrology (Packaged Commodities) Rules, 2011** (including 2017 Amendments).

> **This service does NOT perform OCR or ML.** It receives already-extracted structured label data as JSON and returns a structured violation report. Every violation cites its exact legal clause for full legal traceability.

---

## Ruleset Version

`2011-r6-v1.0` — enforces Rules 2(l), 6, 7, 9, and 11 of the Legal Metrology (Packaged Commodities) Rules, 2011, with 2017 Amendment provisions.

---

## Quick Start

```bash
cd rule_engine
pip install -r requirements.txt
uvicorn main:app --reload
```

- API base: `http://localhost:8000`
- Interactive Swagger UI: `http://localhost:8000/docs`
- ReDoc: `http://localhost:8000/redoc`

```bash
# Run all tests
pytest tests/ -v
```

---

## Folder Structure

```
rule_engine/
├── main.py                     # FastAPI app — POST /validate, GET /rules, GET /health
├── schemas.py                  # Pydantic v2 models for all input/output contracts
├── applicability.py            # Decides which rules run (exemptions, food advisory, import)
├── rule_registry.py            # Maps rule_key → Rule instance; version metadata
├── rules_config.json           # Single source of truth for all thresholds and phrases
├── engine.py                   # Orchestrator: tag bridge → bulk check → rules → response
├── tag_bridge.py               # NEW: extracted_tags → ExtractedField synthesis module
├── conftest.py                 # pytest sys.path configuration
├── requirements.txt
├── rules/
│   ├── __init__.py
│   ├── base.py                 # Abstract Rule interface (rule_id, clause, severity, check())
│   ├── manufacturer_rule.py    # Rule 6(1)(a) — manufacturer/packer/importer
│   ├── net_quantity_rule.py    # Rule 6(1)(c) & Rule 11 — net quantity in standard units
│   ├── mrp_rule.py             # Rule 6(1)(e) — MRP inclusive of all taxes
│   ├── mfg_date_rule.py        # Rule 6(1)(d) — month and year of manufacture
│   ├── consumer_care_rule.py   # Rule 6(1)(f) — consumer helpline (phone/email)
│   ├── country_of_origin_rule.py  # Rule 6(1)(b) proviso — imported goods only
│   ├── font_size_rule.py       # Rule 7 — net quantity numeral height (1mm/2mm)
│   ├── placement_rule.py       # Rule 9 — principal display panel placement
│   └── generic_name_rule.py    # Rule 6(1)(a) & Rule 2(l) — generic/common name
└── tests/
    ├── __init__.py
    ├── test_fixtures.py             # 27 synthetic test cases covering all edge cases
    ├── test_manufacturer_rule.py
    ├── test_mrp_rule.py
    ├── test_font_size_rule.py
    ├── test_applicability.py
    └── test_engine_integration.py   # Full end-to-end + tag bridge tests
```

---

## Rules Reference

| Rule File | Legal Clause | What It Checks | Severity |
|---|---|---|---|
| `manufacturer_rule.py` | Rule 6(1)(a) | Manufacturer/packer name & address; separate `importer` field for imported goods | major |
| `generic_name_rule.py` | Rule 6(1)(a) & Rule 2(l) | Generic or common name of the commodity (brand name alone is insufficient) | major |
| `net_quantity_rule.py` | Rule 6(1)(c) & Rule 11 | Net quantity present, numeric value, valid metric unit (g/kg/ml/l/pcs…); multi-pack aware | major |
| `mrp_rule.py` | Rule 6(1)(e) | MRP present, numeric price, and accepted "incl. of all taxes" phrase | major |
| `mfg_date_rule.py` | Rule 6(1)(d) | Month + year of manufacture/packing; mandatory on physical label for ALL sale channels | major |
| `consumer_care_rule.py` | Rule 6(1)(f) | Phone (10-digit / STD / 1800 toll-free) or email address present | minor |
| `country_of_origin_rule.py` | Rule 6(1)(b) proviso | Country of origin declared; **only runs when `is_imported=True`** | major |
| `font_size_rule.py` | Rule 7 | Net quantity numeral height ≥ 1.0 mm (printed) / ≥ 2.0 mm (embossed/moulded/blown) | minor |
| `placement_rule.py` | Rule 9 | Mandatory fields must appear on the `principal_display_panel` | minor |

---

## API Reference

### `POST /validate`

Run compliance check on extracted label fields.

#### OCR / ML Service Integration — Two Input Modes

The engine accepts input in **two modes** that can be combined freely:

**Mode 1 — `extracted_fields` (structured):** Pass a list of `ExtractedField` objects with bounding boxes, confidence scores, and panel locations. This is the most precise input.

**Mode 2 — `extracted_tags` (flat dict):** Pass the raw tag dictionary from the upstream ML service (e.g., `MANUFACTURER_NAME`, `NET_QUANTITY`, `MRP`, etc.). The engine's **tag bridge** automatically synthesizes `ExtractedField` objects from these tags.

**Both modes can be combined.** Caller-provided `extracted_fields` always take priority — the bridge only fills field_types not already covered.

```
OCR / ML service output
        │
        ├── extracted_tags  ─────────────── tag_bridge.py ──────────▶ synthesized ExtractedFields
        │   {MANUFACTURER_NAME, MRP,                                   (fills gaps only)
        │    NET_QUANTITY, IMPORTER_NAME, …}                                    │
        │                                                                        ▼
        └── extracted_fields  ──────────────────────────────────────▶ merged all_fields
            [{field_type, text, confidence, bbox, …}]                (caller fields first)
                                                                              │
                                                                              ▼
                                                                    Rule engine → ValidationResponse
```

#### Request Body — `ValidationRequest`

| Field | Type | Required | Description |
|---|---|---|---|
| `product_id` | `str` | ✅ | Unique identifier for this product/batch |
| `category` | `str` | ✅ | `food` \| `cosmetics` \| `electronics` \| `other` \| `cement` \| `fertilizer` \| `agricultural_produce` |
| `package_shape` | `str` | ✅ | `rectangular` \| `cylindrical` \| `irregular` — **case-insensitive** |
| `package_surface_type` | `str` | ✅ | `printed` \| `embossed_molded_blown` — determines Rule 7 font threshold |
| `sale_channel` | `str` | ✅ | `retail` \| `ecommerce` |
| `net_weight_kg` | `float` | ❌ | Net weight in kg — triggers bulk exemption if above threshold |
| `net_volume_l` | `float` | ❌ | Net volume in litres — triggers volume-based bulk exemption |
| `is_imported` | `bool` | ✅ | `true` for imported goods — activates `country_of_origin` rule |
| `extracted_fields` | `List[ExtractedField]` | ❌ | Structured fields with bbox/confidence (default `[]`) |
| `extracted_tags` | `Dict[str, Any]` | ❌ | Flat tag dict from upstream ML service — auto-synthesized into fields |
| `text_measurements` | `List[TextMeasurementEntry]` | ❌ | Per-block font height data — used by tag bridge to attach `font_height_mm` |

> Any other keys in the JSON payload (e.g. `panel_analysis`, `image_base64`, `scale_reference_mm`) are silently dropped via `extra="ignore"`. The OCR service can POST its full output without stripping.

#### `ExtractedField` Object

| Field | Type | Description |
|---|---|---|
| `field_type` | `str` | `manufacturer` \| `generic_name` \| `net_quantity` \| `mrp` \| `mfg_date` \| `consumer_care` \| `country_of_origin` \| `importer` |
| `text` | `str` | Raw extracted text |
| `confidence` | `float` | 0–1; fields below `0.75` go to `manual_review_required` |
| `bbox` | `{x, y, w, h}` | Bounding box on the label image |
| `font_height_mm` | `float \| null` | Measured font height in mm — used by Rule 7 for `net_quantity` only |
| `panel_location` | `str` | `principal_display_panel` \| `other` |

#### `extracted_tags` — Recognised Tag Keys

| Tag Key(s) | → `field_type` | Notes |
|---|---|---|
| `MANUFACTURER_NAME` + `MANUFACTURER_ADDRESS` | `manufacturer` | Primary entry |
| `PACKER_NAME` + `PACKER_ADDRESS` | `manufacturer` | Second entry if distinct from manufacturer |
| `IMPORTER_NAME` + `IMPORTER_ADDRESS` | `importer` | Only synthesized when value is non-null |
| `PRODUCT_COMMON_NAME` | `generic_name` | |
| `NET_QUANTITY` (+ `NET_QUANTITY_UNIT`) | `net_quantity` | Unit appended if not already embedded |
| `MRP` + `MRP_INCLUSIVE_OF_TAXES` | `mrp` | Combined into one field text |
| `MANUFACTURE_MONTH` + `MANUFACTURE_YEAR` | `mfg_date` | Joined with space |
| `CONSUMER_CARE_PHONE` + `CONSUMER_CARE_EMAIL` + `CONSUMER_CARE_NAME` | `consumer_care` | All combined |
| `COUNTRY_OF_ORIGIN` | `country_of_origin` | |

#### Example — Full OCR Payload (Tags + Fields Combined)

```bash
curl -X POST http://localhost:8000/validate \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "PARACHUTE-300ML-001",
    "category": "other",
    "package_shape": "CYLINDRICAL",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.273,
    "is_imported": false,
    "extracted_fields": [],
    "extracted_tags": {
      "MANUFACTURER_NAME": "Marico Ltd.",
      "MANUFACTURER_ADDRESS": "7th Floor, Grande Palladium, 175 CST Road, Mumbai - 400 098",
      "PACKER_NAME": "Sowparnika Industries Private Limited",
      "PACKER_ADDRESS": "P.V.Palayam, Perundurai, Erode - 638 052, TN",
      "IMPORTER_NAME": null,
      "IMPORTER_ADDRESS": null,
      "COUNTRY_OF_ORIGIN": "India",
      "PRODUCT_COMMON_NAME": "100% Pure Coconut Oil",
      "NET_QUANTITY": "300 ml + 30 ml FREE",
      "NET_QUANTITY_UNIT": "ml",
      "MANUFACTURE_MONTH": "MAY",
      "MANUFACTURE_YEAR": "2026",
      "MRP": "₹222",
      "MRP_INCLUSIVE_OF_TAXES": "Incl. of all taxes",
      "CONSUMER_CARE_PHONE": "1800222248",
      "CONSUMER_CARE_EMAIL": "csc@marico.com"
    },
    "text_measurements": [
      {"text": "300 ml", "confidence": 0.97, "height_mm": 3.15, "min_required_mm": 1.0, "is_rule7_compliant": true},
      {"text": "Marico Ltd.", "confidence": 0.96, "height_mm": 2.1, "min_required_mm": 1.0, "is_rule7_compliant": true}
    ]
  }'
```

> **Note:** Any additional keys (`panel_analysis`, `text_overlap_analysis`, `scale_reference_mm`, etc.) are silently ignored.

#### Example — Structured `extracted_fields` Only

```bash
curl -X POST http://localhost:8000/validate \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "SKU-RETAIL-001",
    "category": "other",
    "package_shape": "rectangular",
    "package_surface_type": "printed",
    "sale_channel": "retail",
    "net_weight_kg": 0.5,
    "is_imported": false,
    "extracted_fields": [
      {"field_type": "manufacturer", "text": "Acme Industries Pvt. Ltd., 123 Road, Mumbai - 400001", "confidence": 0.95, "bbox": {"x":0,"y":0,"w":100,"h":20}, "font_height_mm": 1.5, "panel_location": "principal_display_panel"},
      {"field_type": "generic_name", "text": "Refined Sunflower Oil", "confidence": 0.97, "bbox": {"x":0,"y":25,"w":100,"h":20}, "font_height_mm": 1.5, "panel_location": "principal_display_panel"},
      {"field_type": "net_quantity", "text": "500 ml", "confidence": 0.98, "bbox": {"x":0,"y":50,"w":100,"h":20}, "font_height_mm": 1.5, "panel_location": "principal_display_panel"},
      {"field_type": "mrp", "text": "MRP Rs. 120/- (Incl. of all taxes)", "confidence": 0.99, "bbox": {"x":0,"y":75,"w":100,"h":20}, "font_height_mm": 1.5, "panel_location": "principal_display_panel"},
      {"field_type": "mfg_date", "text": "Jan 2024", "confidence": 0.96, "bbox": {"x":0,"y":100,"w":100,"h":20}, "font_height_mm": 1.5, "panel_location": "principal_display_panel"},
      {"field_type": "consumer_care", "text": "1800-123-4567", "confidence": 0.94, "bbox": {"x":0,"y":125,"w":100,"h":20}, "font_height_mm": 1.5, "panel_location": "principal_display_panel"}
    ]
  }'
```

#### Response — `ValidationResponse`

```json
{
  "product_id": "PARACHUTE-300ML-001",
  "ruleset_version": "2011-r6-v1.0",
  "overall_status": "compliant",
  "exempt_reason": null,
  "violations": [],
  "manual_review_required": [],
  "applicability_notes": [...],
  "total_violations": 0,
  "major_violations": 0,
  "minor_violations": 0,
  "advisory_violations": 0,
  "synthesized_fields_summary": [
    {
      "field_type": "manufacturer",
      "text": "Marico Ltd., 7th Floor, Grande Palladium...",
      "source_tags": ["MANUFACTURER_NAME", "MANUFACTURER_ADDRESS"],
      "font_height_mm": 2.1,
      "confidence": 0.88
    },
    {
      "field_type": "net_quantity",
      "text": "300 ml + 30 ml FREE",
      "source_tags": ["NET_QUANTITY", "NET_QUANTITY_UNIT"],
      "font_height_mm": 3.15,
      "confidence": 0.88
    }
  ]
}
```

#### Non-Compliant Response

```json
{
  "product_id": "SKU-BAD-002",
  "ruleset_version": "2011-r6-v1.0",
  "overall_status": "non_compliant",
  "exempt_reason": null,
  "violations": [
    {
      "rule_id": "MRP_001",
      "clause": "Rule 6(1)(e)",
      "severity": "major",
      "description": "MRP field 'MRP Rs. 30/-' is missing the mandatory tax-inclusive phrase.",
      "field_evidence": "MRP Rs. 30/-",
      "field_type": "mrp"
    }
  ],
  "manual_review_required": [],
  "applicability_notes": [...],
  "total_violations": 1,
  "major_violations": 1,
  "minor_violations": 0,
  "advisory_violations": 0,
  "synthesized_fields_summary": []
}
```

#### `overall_status` Values

| Value | Meaning |
|---|---|
| `compliant` | All applicable rules passed |
| `non_compliant` | One or more violations found |
| `exempt` | Package exceeds bulk weight/volume threshold — Chapter II does not apply |
| `requires_manual_review` | No violations but some fields are low-confidence — human inspection required |

#### Violation Severity Values

| Severity | Meaning |
|---|---|
| `major` | Prosecution-level violation — directly deceptive to consumers |
| `minor` | Advisory violation — non-conformance but lower legal risk |
| `advisory` | Food-category downgrade — also governed by FSSAI |

> **Invariant:** `major + minor + advisory == total_violations` always holds.

---

### `GET /rules`

Returns all active rules with legal citations, severity levels, and the ruleset version string.

```bash
curl http://localhost:8000/rules
```

---

### `GET /health`

Basic liveness probe for load balancers.

```bash
curl http://localhost:8000/health
```

---

## Tag Bridge (`tag_bridge.py`)

The tag bridge converts the flat `extracted_tags` dictionary from the upstream ML/OCR service into `ExtractedField` objects. It runs as **Step 0** of the engine pipeline, before any rule logic.

### Merging Strategy

```
existing extracted_fields (caller-provided, high confidence)
         ↓ takes priority
tag_bridge.synthesize_from_tags(extracted_tags, text_measurements, existing_fields)
         ↓ fills gaps only — never overwrites caller fields
merged all_fields → confidence split → rule execution
```

### Font Height Resolution

When `text_measurements` are provided, the bridge performs a **substring match** to find the font height for each synthesized field:

```python
# Example: net_quantity text = "300 ml + 30 ml FREE"
# Probe: first 30 chars → "300 ml + 30 ml FREE"[:30]
# Scans text_measurements for any entry whose .text contains the probe
# Match found: {"text": "300 ml", "height_mm": 3.15} → font_height_mm = 3.15
```

If no match is found, `font_height_mm` defaults to `None` (font check is skipped for that field — cannot assess without a measurement).

### Synthesis Confidence

Synthesized fields receive confidence `0.88` by default — above the `0.75` review threshold, meaning they participate fully in rule checks. This reflects that structured tags from the ML service are post-processed outputs, not raw OCR confidence scores.

---

## Applicability Logic

Rules are applied conditionally **before** any compliance checking runs.

### Bulk Package Exemption (Rule 3)

A package is exempt from **all** Chapter II declarations if it exceeds the threshold for its category. Triggered by **either** weight or volume.

| Category | Weight threshold | Volume threshold |
|---|---|---|
| `food`, `cosmetics`, `electronics`, `other` | > 25 kg | > 25 L |
| `cement` | > 50 kg | > 50 L |
| `fertilizer` | > 50 kg | > 50 L |
| `agricultural_produce` | > 50 kg | > 50 L |

Response: `overall_status = "exempt"`, `exempt_reason` explains which threshold was exceeded.

Both thresholds live in the single `bulk_exemption_thresholds` dict in `rules_config.json` (both `_kg` and `_l` keys in one object).

### Rule 6(10) — E-Commerce (2017 Amendment) ⚠️ Corrected

Rule 6(10) requires that **online listings** of pre-packaged commodities display all mandatory declarations. It **adds** an obligation for the e-commerce platform — it does **NOT** exempt the physical label from any declaration.

> **`mfg_date` remains mandatory on the physical label for e-commerce products.**

This is a commonly misread provision. Our engine correctly enforces it: `sale_channel == "ecommerce"` changes nothing about which physical-label rules run.

### Food Category Advisory

When `category == "food"`, the `mfg_date` and `net_quantity` rules are downgraded to `"advisory"` severity — these fields may also be governed by FSSAI. The checks still run; violations are tagged `"advisory"` instead of `"major"`.

### Country of Origin (Conditional)

The `country_of_origin` rule only runs when `is_imported == true`. Domestic products are never flagged.

### Importer Declaration (Conditional)

The `manufacturer` rule's importer sub-check runs only when `is_imported == true`. It checks for an explicit `field_type == "importer"` entry **first** (highest priority), then falls back to keyword-scanning manufacturer fields.

### Confidence Threshold

Configured in `rules_config.json` as `"confidence_threshold": 0.75`.

- Fields **below** threshold → `manual_review_required` — never auto-judged
- If **all** extractions for a field_type are low-confidence → rule is **skipped entirely** (no spurious "field missing" violation)
- Synthesized tag-bridge fields → confidence `0.88` (above threshold, participate in checks)

---

## Rule 7 — Correct Legal Scope ⚠️

Rule 7 ("Numeral size") prescribes minimum character heights **only for the net quantity numeral**. Other mandatory declarations (manufacturer, MRP, generic name, consumer care, mfg_date) are governed by Rule 6's qualitative "legible and conspicuous" standard — **no statutory mm figure is attached to those fields**.

| Field | Rule 7 applies? | Threshold |
|---|---|---|
| `net_quantity` | ✅ Yes | 1.0 mm (printed), 2.0 mm (embossed) |
| `mrp` | ❌ No | "legible" — qualitative, Rule 6 |
| `manufacturer` | ❌ No | "legible" — qualitative, Rule 6 |
| `generic_name` | ❌ No | "legible" — qualitative, Rule 6 |
| `mfg_date` | ❌ No | "legible" — qualitative, Rule 6 |
| `consumer_care` | ❌ No | "legible" — qualitative, Rule 6 |

---

## `rules_config.json` — Configuration Reference

Single source of truth for all thresholds and phrases. Editing this file changes compliance behaviour without touching rule code.

| Key | Purpose |
|---|---|
| `ruleset_version` | Semantic version string returned in every response |
| `confidence_threshold` | Float 0–1; fields below this go to `manual_review_required` |
| `bulk_exemption_thresholds` | Single flat dict with both `_kg` and `_l` keys per category |
| `font_size_thresholds_mm` | Minimum numeral heights for `printed` and `embossed_molded_blown` |
| `mrp_required_phrases` | Accepted variants of the "incl. of all taxes" phrase |
| `net_quantity_allowed_units` | Allowed metric unit strings grouped by weight/volume/length/count/area |
| `consumer_care_patterns` | Regex for valid Indian phone numbers and email addresses |
| `mfg_date_patterns` | Regex patterns for accepted month-year date formats |
| `mandatory_field_types` | Field types checked by placement rule |
| `placement_mandatory_on_principal_panel` | Field types that must appear on the principal display panel |
| `_ecommerce_note` | Documents correct Rule 6(10) interpretation (not a config value) |

---

## MRP Accepted Phrases

The following satisfy Rule 6(1)(e) (all case-insensitive):

- `incl. of all taxes`
- `inclusive of all taxes`
- `incl of all taxes`
- `incl. of taxes`
- `incl of taxes`
- `including all taxes`
- `all taxes inclusive`
- `max. retail price (incl. of all taxes)`
- `mrp (incl. of all taxes)`
- `mrp incl of all taxes`

To add a phrase: edit `mrp_required_phrases` in `rules_config.json`. No code change required.

---

## Net Quantity — Unit Parsing

The net quantity parser (`net_quantity_rule.py`) handles:

| Format | Example | Detected unit |
|---|---|---|
| Simple | `500 g`, `1.5 kg`, `200 ml` | `g`, `kg`, `ml` |
| Multi-pack | `2 x 325 ml`, `6 x 100 g` | `ml`, `g` (from per-unit part) |
| Bonus-pack | `300 ml + 30 ml FREE`, `500 g + 50 g free` | `ml`, `g` (from base quantity) |
| Compound area | `2 sq.m`, `5 sq ft` | `sqm`, `sqft` |
| Count | `12 pcs`, `1 nos` | `pcs`, `nos` |

The multi-pack pre-check runs **first** — preventing `x` from being tokenised as a unit.

---

## Manufacturer / Importer Rule Logic

`MANUFACTURER_001` checks importer declaration in **three-tier priority order**:

| Priority | What is checked |
|---|---|
| ① (highest) | Explicit `field_type == "importer"` entry in `extracted_fields` or synthesized from `IMPORTER_NAME` / `IMPORTER_ADDRESS` tags |
| ② | Importer keywords (`import`, `imported by`, `marketed by`, `distributor`) in any `manufacturer`-type field text |
| ③ | Multiple `manufacturer`-type entries (second likely represents importer) |

If none of the three are satisfied and `is_imported == true`, `MANUFACTURER_001` fires with `field_type = "importer"` in the violation.

---

## How to Add a New Rule

1. **Create `rules/my_new_rule.py`** — implement `check()` as a pure function of `(fields, context)`.

2. **Export from `rules/__init__.py`**.

3. **Register in `rule_registry.py`**:
```python
from rules.my_new_rule import MyNewRule
RULE_REGISTRY["my_new_rule"] = MyNewRule()
```

4. **Add to `applicability.py` `ALL_RULE_KEYS`**:
```python
ALL_RULE_KEYS = frozenset([..., "my_new_rule"])
```

5. **Add config entry** (if the rule needs thresholds) to `rules_config.json`.

6. **Add to `tag_bridge.py`** (if the rule checks a field that comes from a new tag key).

7. **Write tests** in `tests/test_my_new_rule.py` + add a fixture to `tests/test_fixtures.py`.

---

## Test Coverage

```
tests/test_fixtures.py            — 27 annotated synthetic test cases
tests/test_manufacturer_rule.py   — presence, text length, importer check (3-tier priority)
tests/test_mrp_rule.py            — presence, 9+ phrase variants, numeric price
tests/test_font_size_rule.py      — printed/embossed thresholds, Rule 7 scope (net_quantity only)
tests/test_applicability.py       — bulk weight/volume exemption, food advisory, import flag
tests/test_engine_integration.py  — full end-to-end + tag bridge (TestTagBridge)
```

```bash
pytest tests/ -v
```

---

## Engine Pipeline

```
POST /validate
     │
     ├─ Step 0: Tag bridge (tag_bridge.py)
     │      extracted_tags → ExtractedField synthesis
     │      text_measurements → font_height_mm lookup
     │      caller extracted_fields + synthesized → all_fields
     │
     ├─ Step 1: Bulk exemption (applicability.py)
     │      → if exempt: return immediately with status="exempt"
     │
     ├─ Step 2: Determine applicable rules (applicability.py)
     │      → food advisory downgrades, import flag, bulk-category check
     │
     ├─ Step 3: Confidence split (engine.py)
     │      → high-confidence fields  → passed to rules
     │      → low-confidence types    → manual_review_required
     │         (only if NO high-confidence extraction exists for that type)
     │
     ├─ Step 4: Run each applicable rule (rule_registry.py → rules/*.py)
     │      → each rule returns None (pass) or ViolationResult (fail)
     │      → severity may be overridden (e.g. major → advisory for food)
     │
     ├─ Step 5: Determine overall_status
     │      violations?            → "non_compliant"
     │      manual_review only?    → "requires_manual_review"
     │      neither?               → "compliant"
     │
     └─ Step 6: Assemble and return ValidationResponse
                includes synthesized_fields_summary for audit trail
```

---

## Design Principles

| Principle | How It's Enforced |
|---|---|
| **Zero ML in rule logic** | Every compliance decision is an explicit `if/else` against a quoted statute clause |
| **OCR-envelope compatible** | `extra="ignore"` on `ValidationRequest` — upstream service can POST its full output |
| **Tag bridge for accuracy** | `extracted_tags` synthesized automatically — works with or without `extracted_fields` |
| **Single config source** | All thresholds and phrases live in `rules_config.json`; no magic numbers in rule code |
| **Legal traceability** | Every `ViolationResult` carries `rule_id`, `clause`, `severity`, `description` |
| **Versioned ruleset** | `ruleset_version` returned in every response for audit trail |
| **Pure rule functions** | Each rule's `check()` is a pure function — no shared mutable state |
| **Confidence-aware** | Low-confidence extractions → `manual_review_required`, never auto-judged |
| **Correct Rule 7 scope** | Font size mm threshold applies to net_quantity only (not all fields) |
| **Correct Rule 6(10) scope** | E-commerce does NOT exempt physical label from mfg_date or any other field |

---

## Legal References

- Legal Metrology (Packaged Commodities) Rules, 2011 — Ministry of Consumer Affairs, Food and Public Distribution
- Legal Metrology (Packaged Commodities) Amendment Rules, 2017
- Country of Origin (Compulsory Indication) Order, 1983
- Food Safety and Standards Authority of India (FSSAI) Regulations (food category cross-reference)
