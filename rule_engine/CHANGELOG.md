# CHANGELOG

All notable changes to the **Legal Metrology Rule Engine** are documented here.  
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] — v1.1.0 (tag bridge integration branch)

### Added
- **`tag_bridge.py`** — new module that converts `extracted_tags` + `text_measurements`
  from the upstream OCR/ML service into `ExtractedField` objects automatically.
  - Covers all 7 mandatory declarations: manufacturer, packer, importer, generic_name,
    net_quantity, mfg_date, mrp, consumer_care, country_of_origin.
  - Caller-provided `extracted_fields` always win; bridge only fills field_type gaps.
  - Font height resolved via **closest-length substring match** against `text_measurements`
    (order-independent; exact matches always preferred over partial substring matches).
  - Synthesized fields receive confidence `0.88` — above the 0.75 threshold, fully
    participating in all rule checks.
- **`schemas.py`** — `TextMeasurementEntry` model; `extracted_tags` and `text_measurements`
  optional fields on `ValidationRequest`; `synthesized_fields_summary` on `ValidationResponse`.
- **`engine.py`** — Step 0 tag bridge integration: `all_fields = extracted_fields + synthesized`.
- **`rules/generic_name_rule.py`** — Rule 6(1)(a) & Rule 2(l) generic/common name check
  (previously a phantom rule registered in `ALL_RULE_KEYS` but never implemented).
- **`tests/test_tag_bridge_unit.py`** — 23 unit tests for `_find_font_height` and
  `synthesize_from_tags` (exact match, ambiguous order, no-match null, short-token guard,
  merge priority, null tag handling).
- **`tests/test_engine_integration.py`** — `TestTagBridge` (8 tests) and `TestAPIDemoRegressions`
  (15 tests) classes derived from real API call observations.
- **`tests/test_fixtures.py`** — Fixtures 24–32 covering all tag bridge scenarios.

### Fixed
- **`MANUFACTURER_001` importer blind spot**: rule now checks for an explicit
  `field_type == "importer"` entry (priority ①) before keyword-scanning manufacturer
  field text. Audit trail correctly reports `field_type = "importer"` in violation.
- **`FONT_SIZE_001` over-broad scope**: Rule 7's numeric mm threshold now applies only
  to `net_quantity` numerals. Other mandatory fields (mrp, manufacturer, etc.) fall under
  Rule 6's qualitative "legible" standard — no mm figure attached.
- **`NET_QUANTITY_001` multi-pack false positive**: added Path 0 pre-check recognising
  `N x qty unit` and `qty unit + qty unit FREE` bonus-pack notation. `"2 x 325 ml"` and
  `"300 ml + 30 ml FREE"` now correctly extract unit `ml` instead of tokenising on `x`.
- **Rule 6(10) e-commerce interpretation**: removed incorrect exemption of `mfg_date` for
  e-commerce products. Rule 6(10) 2017 Amendment adds an online-listing obligation; it does
  NOT remove any declaration from the physical label. All tests and documentation updated.
- **`_find_font_height` substring direction bug**: old code checked if the full field text
  was a substring of the (shorter) measurement text — always False. Fixed to check if the
  measurement text is a substring of the field text.
- **`_find_font_height` order-dependence**: upgraded from first-match to closest-length-match
  algorithm. Exact matches always win; resolution is now order-independent (SPICE-TAGS-002).
- Stale `"ecommerce exempt"` example removed from `applicability.py` docstring.
- Stale `test_ecommerce_mfg_date_excluded` test corrected to assert `apply is True`.
- `test_font_fail_1_0mm_embossed` and `test_multiple_fields_failing` updated to reflect
  correct Rule 7 scope (net_quantity only; mrp not checked by Rule 7).

### Changed
- `ValidationRequest` now accepts full OCR service output via `extra="ignore"` — all
  unknown fields (panel_analysis, image_base64, scale_reference_mm, etc.) silently dropped.
- `rules_config.json`: removed stale `ecommerce_exempt_fields` key; added `_ecommerce_note`
  documenting correct Rule 6(10) interpretation.
- **Confidence-based field selection clarified** (design tradeoff, not a bug): when multiple
  extractions exist for a field_type, only high-confidence (≥0.75) entries are passed to
  rules. If the correct text happens to be on a lower-confidence extraction, the high-conf
  text is used and may produce a violation. Documented in BISCUIT-TAGS-005 fixture.

---

## [1.0.0] — 2026-09-01

### Added
- Initial rule engine implementation.
- Rules: MANUFACTURER_001, NET_QUANTITY_001, MRP_001, MFG_DATE_001, CONSUMER_CARE_001,
  COUNTRY_ORIGIN_001, FONT_SIZE_001, PLACEMENT_001.
- Bulk exemption by weight (Rule 3) for categories: other, cement, fertilizer,
  agricultural_produce.
- Volume-based bulk exemption (`net_volume_l`).
- Food-category advisory downgrade for mfg_date and net_quantity.
- Country-of-origin conditional rule (imported goods only).
- Confidence threshold (0.75): low-confidence fields → `manual_review_required`.
- Ruleset versioning via `rules_config.json`.
- FastAPI app with `/validate`, `/rules`, `/health` endpoints.
- 23 integration test fixtures.
