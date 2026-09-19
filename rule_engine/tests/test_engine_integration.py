"""
tests/test_engine_integration.py

End-to-end integration tests for engine.validate().
Runs full validation pipeline and checks complete response shapes.
"""

from __future__ import annotations

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from schemas import ValidationRequest, ValidationResponse
from engine import validate
from tests.test_fixtures import (
    FIXTURE_COMPLIANT_RETAIL,
    FIXTURE_MISSING_MRP,
    FIXTURE_MRP_MISSING_TAX_PHRASE,
    FIXTURE_FONT_FAIL_PRINTED,
    FIXTURE_FONT_FAIL_EMBOSSED,
    FIXTURE_MISSING_CONSUMER_CARE,
    FIXTURE_ECOMMERCE_NO_MFG_DATE,
    FIXTURE_IMPORTED_MISSING_COUNTRY,
    FIXTURE_DOMESTIC_NO_COUNTRY_REQUIRED,
    FIXTURE_EXEMPT_BULK,
    FIXTURE_EXEMPT_CEMENT_BULK,
    FIXTURE_FOOD_ADVISORY,
    FIXTURE_LOW_CONFIDENCE_CONSUMER_CARE,
    FIXTURE_WRONG_PANEL_PLACEMENT,
    FIXTURE_MULTIPLE_VIOLATIONS,
    FIXTURE_IMPORTED_WITH_COUNTRY,
    FIXTURE_INVALID_QUANTITY_UNIT,
    FIXTURE_CONSUMER_CARE_NO_CONTACT,
    FIXTURE_EXEMPT_VOLUME,
    FIXTURE_MISSING_GENERIC_NAME,
    # Tag bridge fixtures
    FIXTURE_TAGS_ONLY_DOMESTIC,
    FIXTURE_TAGS_MERGE_WITH_FIELDS,
    FIXTURE_TAGS_IMPORTED_WITH_IMPORTER,
    FIXTURE_TAGS_MULTIPACK,
    # API demo regression fixtures
    FIXTURE_BISCUIT_DUAL_CONFIDENCE_MRP,
    FIXTURE_SPICE_AMBIGUOUS_FONT_MEASUREMENT,
    FIXTURE_CANDLE_NO_MEASUREMENT_MATCH,
    FIXTURE_COOKINGOIL_IMPORTED_BONUS_PACK,
    FIXTURE_SHAMPOO_ECOMMERCE_MISSING_MFGDATE,
)


def _run(fixture: dict) -> ValidationResponse:
    request = ValidationRequest(**fixture)
    return validate(request)


class TestResponseShape:
    """Validates that the response always has the correct schema."""

    def test_response_has_required_fields(self):
        """Every response must have product_id, ruleset_version, overall_status, etc."""
        resp = _run(FIXTURE_COMPLIANT_RETAIL)
        assert hasattr(resp, "product_id")
        assert hasattr(resp, "ruleset_version")
        assert hasattr(resp, "overall_status")
        assert hasattr(resp, "violations")
        assert hasattr(resp, "manual_review_required")
        assert hasattr(resp, "applicability_notes")
        assert hasattr(resp, "total_violations")
        assert hasattr(resp, "major_violations")
        assert hasattr(resp, "minor_violations")

    def test_ruleset_version_populated(self):
        """Ruleset version must match the config."""
        resp = _run(FIXTURE_COMPLIANT_RETAIL)
        assert resp.ruleset_version.startswith("2011")

    def test_product_id_passes_through(self):
        """product_id in response must match input."""
        resp = _run(FIXTURE_COMPLIANT_RETAIL)
        assert resp.product_id == "TEST-001"

    def test_violation_has_required_fields(self):
        """Each violation must have rule_id, clause, severity, description."""
        resp = _run(FIXTURE_MISSING_MRP)
        assert len(resp.violations) > 0
        v = resp.violations[0]
        assert v.rule_id
        assert v.clause
        assert v.severity in ("major", "minor", "advisory")
        assert v.description


class TestCompliantProduct:
    """A fully compliant product must return clean."""

    def test_compliant_retail_status(self):
        resp = _run(FIXTURE_COMPLIANT_RETAIL)
        assert resp.overall_status == "compliant"
        assert resp.violations == []
        assert resp.total_violations == 0
        assert resp.exempt_reason is None


class TestViolationDetection:
    """Each violation scenario produces the correct rule_id."""

    def test_missing_mrp(self):
        resp = _run(FIXTURE_MISSING_MRP)
        assert resp.overall_status == "non_compliant"
        rule_ids = [v.rule_id for v in resp.violations]
        assert "MRP_001" in rule_ids

    def test_mrp_missing_tax_phrase(self):
        resp = _run(FIXTURE_MRP_MISSING_TAX_PHRASE)
        assert resp.overall_status == "non_compliant"
        mrp_violations = [v for v in resp.violations if v.rule_id == "MRP_001"]
        assert len(mrp_violations) > 0
        assert any("tax" in v.description.lower() for v in mrp_violations)

    def test_font_fail_printed(self):
        resp = _run(FIXTURE_FONT_FAIL_PRINTED)
        assert resp.overall_status == "non_compliant"
        font_violations = [v for v in resp.violations if v.rule_id == "FONT_SIZE_001"]
        assert len(font_violations) > 0

    def test_font_fail_embossed(self):
        resp = _run(FIXTURE_FONT_FAIL_EMBOSSED)
        assert resp.overall_status == "non_compliant"
        font_violations = [v for v in resp.violations if v.rule_id == "FONT_SIZE_001"]
        assert len(font_violations) > 0
        assert any("2.0" in v.description for v in font_violations)

    def test_missing_consumer_care(self):
        resp = _run(FIXTURE_MISSING_CONSUMER_CARE)
        assert resp.overall_status == "non_compliant"
        rule_ids = [v.rule_id for v in resp.violations]
        assert "CONSUMER_CARE_001" in rule_ids

    def test_imported_missing_country(self):
        resp = _run(FIXTURE_IMPORTED_MISSING_COUNTRY)
        assert resp.overall_status == "non_compliant"
        rule_ids = [v.rule_id for v in resp.violations]
        assert "COUNTRY_ORIGIN_001" in rule_ids

    def test_wrong_panel_placement(self):
        resp = _run(FIXTURE_WRONG_PANEL_PLACEMENT)
        assert resp.overall_status == "non_compliant"
        placement_violations = [v for v in resp.violations if v.rule_id == "PLACEMENT_001"]
        assert len(placement_violations) > 0
        assert any("net_quantity" in (v.field_type or "") for v in placement_violations)

    def test_invalid_quantity_unit(self):
        resp = _run(FIXTURE_INVALID_QUANTITY_UNIT)
        assert resp.overall_status == "non_compliant"
        rule_ids = [v.rule_id for v in resp.violations]
        assert "NET_QUANTITY_001" in rule_ids

    def test_consumer_care_no_contact(self):
        resp = _run(FIXTURE_CONSUMER_CARE_NO_CONTACT)
        assert resp.overall_status == "non_compliant"
        rule_ids = [v.rule_id for v in resp.violations]
        assert "CONSUMER_CARE_001" in rule_ids

    def test_multiple_violations_count(self):
        resp = _run(FIXTURE_MULTIPLE_VIOLATIONS)
        assert resp.overall_status == "non_compliant"
        assert resp.total_violations >= 2  # MRP + consumer_care at minimum
        assert resp.total_violations == len(resp.violations)
        # Invariant: all severity buckets sum to total
        assert (
            resp.major_violations + resp.minor_violations + resp.advisory_violations
            == resp.total_violations
        )


class TestExemptScenarios:
    """Bulk-exempt packages return early with status='exempt'."""

    def test_exempt_bulk_package(self):
        resp = _run(FIXTURE_EXEMPT_BULK)
        assert resp.overall_status == "exempt"
        assert resp.exempt_reason is not None
        assert "exempt_bulk_package" in resp.exempt_reason
        assert resp.violations == []

    def test_exempt_cement_bulk(self):
        resp = _run(FIXTURE_EXEMPT_CEMENT_BULK)
        assert resp.overall_status == "exempt"
        assert resp.exempt_reason is not None


class TestEcommerceExemption:
    """
    Rule 6(10) 2017 Amendment — adds an online DISPLAY obligation for e-commerce
    platforms. It does NOT remove any declaration from the physical label.
    mfg_date remains mandatory on the physical package for all sale channels.
    """

    def test_ecommerce_mfg_date_still_required(self):
        """Physical label mfg_date must still be checked for e-commerce products."""
        resp = _run(FIXTURE_ECOMMERCE_NO_MFG_DATE)
        mfg_violations = [v for v in resp.violations if v.rule_id == "MFG_DATE_001"]
        assert len(mfg_violations) > 0, (
            "mfg_date IS required on the physical label even for e-commerce. "
            "Rule 6(10) adds an online-listing obligation — it does not remove "
            "the physical label declaration requirement."
        )

    def test_ecommerce_mfg_date_violation_cites_correct_clause(self):
        """mfg_date violation for e-commerce must still cite Rule 6(1)(d)."""
        resp = _run(FIXTURE_ECOMMERCE_NO_MFG_DATE)
        mfg_v = next((v for v in resp.violations if v.rule_id == "MFG_DATE_001"), None)
        assert mfg_v is not None
        assert mfg_v.clause == "Rule 6(1)(d)"


class TestImportRules:
    """Country of origin — conditional on is_imported."""

    def test_domestic_country_not_required(self):
        resp = _run(FIXTURE_DOMESTIC_NO_COUNTRY_REQUIRED)
        country_violations = [v for v in resp.violations if v.rule_id == "COUNTRY_ORIGIN_001"]
        assert country_violations == [], "Domestic products must NOT require country_of_origin"

    def test_imported_with_country_compliant(self):
        resp = _run(FIXTURE_IMPORTED_WITH_COUNTRY)
        country_violations = [v for v in resp.violations if v.rule_id == "COUNTRY_ORIGIN_001"]
        assert country_violations == []


class TestConfidenceThreshold:
    """Low-confidence fields must go to manual_review_required, NOT violations."""

    def test_low_confidence_field_in_manual_review(self):
        resp = _run(FIXTURE_LOW_CONFIDENCE_CONSUMER_CARE)
        assert "consumer_care" in resp.manual_review_required, (
            "Low-confidence consumer_care should be in manual_review_required"
        )

    def test_low_confidence_field_not_in_violations(self):
        resp = _run(FIXTURE_LOW_CONFIDENCE_CONSUMER_CARE)
        consumer_care_violations = [
            v for v in resp.violations if v.field_type == "consumer_care"
        ]
        assert consumer_care_violations == [], (
            "Low-confidence consumer_care must NOT appear as a violation — "
            "it cannot be judged without human review"
        )

    def test_status_requires_manual_review_when_only_low_conf(self):
        """When only low-confidence fields are missing, status is 'requires_manual_review'."""
        resp = _run(FIXTURE_LOW_CONFIDENCE_CONSUMER_CARE)
        # This fixture has all other fields passing — only consumer_care is low-conf
        assert resp.overall_status in ("compliant", "requires_manual_review")


class TestFoodAdvisory:
    """Food category violations should be tagged with advisory severity."""

    def test_food_compliant_product_passes(self):
        resp = _run(FIXTURE_FOOD_ADVISORY)
        # FIXTURE_FOOD_ADVISORY has all fields properly filled — should be compliant
        assert resp.overall_status == "compliant"

    def test_applicability_notes_show_advisory(self):
        resp = _run(FIXTURE_FOOD_ADVISORY)
        advisory_notes = [
            n for n in resp.applicability_notes
            if n.severity_override == "advisory"
        ]
        assert len(advisory_notes) > 0, "Food category should produce advisory notes"


class TestViolationCitations:
    """Every violation must cite its exact legal clause."""

    def test_all_violations_have_clause(self):
        """Run multiple failure scenarios and confirm every violation has a clause."""
        for fixture in [
            FIXTURE_MISSING_MRP,
            FIXTURE_FONT_FAIL_PRINTED,
            FIXTURE_WRONG_PANEL_PLACEMENT,
            FIXTURE_IMPORTED_MISSING_COUNTRY,
            FIXTURE_CONSUMER_CARE_NO_CONTACT,
        ]:
            resp = _run(fixture)
            for v in resp.violations:
                assert v.clause, f"Violation {v.rule_id} has no clause citation!"
                assert v.rule_id, "Violation has no rule_id!"
                assert v.description, "Violation has no description!"

    def test_mrp_violation_cites_correct_clause(self):
        resp = _run(FIXTURE_MISSING_MRP)
        mrp_v = next((v for v in resp.violations if v.rule_id == "MRP_001"), None)
        assert mrp_v is not None
        assert mrp_v.clause == "Rule 6(1)(e)"

    def test_manufacturer_violation_cites_correct_clause(self):
        fixture = dict(FIXTURE_COMPLIANT_RETAIL)
        fixture["extracted_fields"] = [
            f for f in FIXTURE_COMPLIANT_RETAIL["extracted_fields"]
            if f["field_type"] != "manufacturer"
        ]
        resp = _run(fixture)
        mfr_v = next((v for v in resp.violations if v.rule_id == "MANUFACTURER_001"), None)
        assert mfr_v is not None
        assert mfr_v.clause == "Rule 6(1)(a)"


class TestVolumeExemption:
    """Volume-based bulk exemption must work independently of weight."""

    def test_exempt_by_volume_returns_exempt_status(self):
        """30 L oil drum — no weight given, only volume drives exemption."""
        resp = _run(FIXTURE_EXEMPT_VOLUME)
        assert resp.overall_status == "exempt"
        assert resp.exempt_reason is not None
        assert "exempt_bulk_package" in resp.exempt_reason
        assert "L" in resp.exempt_reason or "litre" in resp.exempt_reason.lower()

    def test_exempt_by_volume_has_no_violations(self):
        """Volume-exempt packages must return empty violations list."""
        resp = _run(FIXTURE_EXEMPT_VOLUME)
        assert resp.violations == []
        assert resp.total_violations == 0

    def test_not_exempt_small_volume(self):
        """10 L — below 25 L threshold — should NOT be exempt."""
        fixture = dict(FIXTURE_EXEMPT_VOLUME)
        fixture["net_volume_l"] = 10.0
        # Need some fields so it doesn't just return compliant vacuously
        fixture["extracted_fields"] = []
        resp = _run(fixture)
        assert resp.overall_status != "exempt"


class TestGenericNameRule:
    """Generic name rule must be real — no longer a phantom."""

    def test_missing_generic_name_fires_violation(self):
        """Missing generic_name field — GENERIC_NAME_001 must fire."""
        resp = _run(FIXTURE_MISSING_GENERIC_NAME)
        assert resp.overall_status == "non_compliant"
        rule_ids = [v.rule_id for v in resp.violations]
        assert "GENERIC_NAME_001" in rule_ids, (
            f"Expected GENERIC_NAME_001 in violations, got: {rule_ids}"
        )

    def test_generic_name_cites_correct_clause(self):
        """Generic name violation must cite Rule 6(1)(a) & Rule 2(l)."""
        resp = _run(FIXTURE_MISSING_GENERIC_NAME)
        gn_v = next((v for v in resp.violations if v.rule_id == "GENERIC_NAME_001"), None)
        assert gn_v is not None
        assert gn_v.clause == "Rule 6(1)(a) & Rule 2(l)"
        assert gn_v.severity == "major"

    def test_compliant_product_with_generic_name_passes(self):
        """Compliant product with generic_name present — no GENERIC_NAME_001 violation."""
        resp = _run(FIXTURE_COMPLIANT_RETAIL)
        gn_violations = [v for v in resp.violations if v.rule_id == "GENERIC_NAME_001"]
        assert gn_violations == [], (
            "Compliant product with generic_name present should not have GENERIC_NAME_001 violation"
        )

    def test_generic_name_rule_is_registered(self):
        """Verify rule registry has the generic_name entry (not phantom)."""
        from rule_registry import RULE_REGISTRY
        assert "generic_name" in RULE_REGISTRY
        rule = RULE_REGISTRY["generic_name"]
        assert rule.rule_id == "GENERIC_NAME_001"
        assert rule.clause == "Rule 6(1)(a) & Rule 2(l)"


class TestTagBridge:
    """
    Tests for the extracted_tags → ExtractedField synthesis path.
    The bridge fills in field_types absent from extracted_fields using
    structured tags from the upstream OCR service.
    """

    def test_tags_only_domestic_is_compliant(self):
        """
        All 7 mandatory fields synthesized from extracted_tags alone.
        No extracted_fields provided. Result must be compliant.
        """
        resp = _run(FIXTURE_TAGS_ONLY_DOMESTIC)
        assert resp.overall_status == "compliant", (
            f"Expected compliant, got {resp.overall_status}. "
            f"Violations: {[v.rule_id for v in resp.violations]}"
        )
        assert resp.violations == []

    def test_tags_only_synthesized_fields_appear_in_summary(self):
        """Response must list all synthesized fields in synthesized_fields_summary."""
        resp = _run(FIXTURE_TAGS_ONLY_DOMESTIC)
        assert len(resp.synthesized_fields_summary) > 0, (
            "No synthesized fields reported — bridge did not run or returned empty"
        )
        field_types = {s["field_type"] for s in resp.synthesized_fields_summary}
        # At minimum: manufacturer, generic_name, net_quantity, mrp, mfg_date, consumer_care
        for expected_type in ("manufacturer", "generic_name", "net_quantity", "mrp", "mfg_date", "consumer_care"):
            assert expected_type in field_types, (
                f"Expected '{expected_type}' in synthesized summary, got: {field_types}"
            )

    def test_tags_summary_records_source_tag_keys(self):
        """Each synthesized entry must record which source tags it was built from."""
        resp = _run(FIXTURE_TAGS_ONLY_DOMESTIC)
        for entry in resp.synthesized_fields_summary:
            assert "source_tags" in entry
            assert isinstance(entry["source_tags"], list)
            assert len(entry["source_tags"]) > 0

    def test_caller_fields_take_priority_over_tags(self):
        """
        Merge test: manufacturer in extracted_fields must NOT be overridden
        by the tag bridge even though MANUFACTURER_NAME is in extracted_tags.
        """
        resp = _run(FIXTURE_TAGS_MERGE_WITH_FIELDS)
        # Manufacturer violation should not fire — caller-provided field is valid
        mfr_violations = [v for v in resp.violations if v.rule_id == "MANUFACTURER_001"]
        assert mfr_violations == [], (
            "Manufacturer from extracted_fields should take priority; "
            "bridge must not overwrite it with tag value"
        )

    def test_tags_ignored_when_field_already_present(self):
        """MANUFACTURER_NAME tag should NOT appear in synthesized_fields_summary
        when a manufacturer extracted_field is already present."""
        resp = _run(FIXTURE_TAGS_MERGE_WITH_FIELDS)
        # The summary should only contain fields synthesized for ABSENT types
        synthesized_types = {s["field_type"] for s in resp.synthesized_fields_summary}
        assert "manufacturer" not in synthesized_types, (
            "manufacturer already in extracted_fields — bridge must not synthesize it again"
        )
        assert "generic_name" not in synthesized_types, (
            "generic_name already in extracted_fields — bridge must not synthesize it again"
        )

    def test_importer_tag_synthesized_for_imported_product(self):
        """
        Imported product: IMPORTER_NAME + IMPORTER_ADDRESS in tags must produce
        a synthesized 'importer' field, and MANUFACTURER_001 must not fire.
        """
        resp = _run(FIXTURE_TAGS_IMPORTED_WITH_IMPORTER)
        mfr_violations = [v for v in resp.violations if v.rule_id == "MANUFACTURER_001"]
        assert mfr_violations == [], (
            f"MANUFACTURER_001 should not fire when importer is declared via tags. "
            f"Got violations: {[v.description for v in mfr_violations]}"
        )
        # Confirm importer was synthesized
        synthesized_types = {s["field_type"] for s in resp.synthesized_fields_summary}
        assert "importer" in synthesized_types, (
            "'importer' field_type must appear in synthesized_fields_summary for imported product"
        )

    def test_multipack_notation_via_tags_passes(self):
        """
        '2 x 325 ml' from NET_QUANTITY tag must be parsed correctly.
        NET_QUANTITY_001 must NOT fire (ml is a valid unit).
        """
        resp = _run(FIXTURE_TAGS_MULTIPACK)
        qty_violations = [v for v in resp.violations if v.rule_id == "NET_QUANTITY_001"]
        assert qty_violations == [], (
            f"'2 x 325 ml' should be recognised as valid multi-pack notation. "
            f"Got: {[v.description for v in qty_violations]}"
        )

    def test_font_height_from_text_measurements_attached(self):
        """
        When text_measurements are provided, the synthesized net_quantity field
        should receive a non-None font_height_mm resolved from the measurements.
        """
        resp = _run(FIXTURE_TAGS_ONLY_DOMESTIC)
        qty_entry = next(
            (s for s in resp.synthesized_fields_summary if s["field_type"] == "net_quantity"),
            None,
        )
        assert qty_entry is not None
        # '300 ml' appears in text_measurements with height_mm=3.15
        assert qty_entry["font_height_mm"] is not None, (
            "font_height_mm should be resolved from text_measurements for net_quantity"
        )
        assert qty_entry["font_height_mm"] == pytest.approx(3.15, abs=0.01)


class TestAPIDemoRegressions:
    """
    Regression tests derived from real API call observations (BISCUIT, SPICE,
    CANDLE, COOKINGOIL, SHAMPOO). Each test pins a specific engine behaviour
    that was verified manually and must not regress.
    """

    # ------------------------------------------------------------------
    # BISCUIT-TAGS-005: dual-confidence MRP
    # ------------------------------------------------------------------

    def test_biscuit_high_conf_mrp_used_low_conf_discarded(self):
        """
        When two mrp fields exist — one low-confidence (0.74, has tax phrase)
        and one high-confidence (0.91, missing tax phrase) — the engine uses
        only the high-confidence field. MRP_001 fires because the high-conf
        text "MRP ₹40/-" lacks the mandatory tax-inclusive phrase.

        This pins the confidence-based selection tradeoff as documented:
        the correct text on the lower-confidence extraction does NOT prevent
        the violation, as per design.
        """
        resp = _run(FIXTURE_BISCUIT_DUAL_CONFIDENCE_MRP)
        mrp_violations = [v for v in resp.violations if v.rule_id == "MRP_001"]
        assert len(mrp_violations) > 0, (
            "High-confidence 'MRP ₹40/-' is missing the tax phrase — "
            "MRP_001 must fire even though a lower-confidence field had correct text"
        )

    def test_biscuit_low_conf_mrp_not_in_manual_review(self):
        """
        Even though mrp has a sub-threshold (0.74) extraction, a HIGH-confidence
        extraction also exists. When any high-confidence extraction exists for a
        field_type, that type must NOT appear in manual_review_required.
        """
        resp = _run(FIXTURE_BISCUIT_DUAL_CONFIDENCE_MRP)
        assert "mrp" not in resp.manual_review_required, (
            "When a high-confidence mrp extraction exists, mrp must NOT be in "
            "manual_review_required — even if a sub-threshold extraction also exists"
        )

    # ------------------------------------------------------------------
    # SPICE-TAGS-002: ambiguous font measurement order
    # ------------------------------------------------------------------

    def test_spice_exact_font_match_wins_over_list_order(self):
        """
        SPICE-TAGS-002: text_measurements has two entries where the longer
        entry comes first in the list. The closest-length algorithm must pick
        the exact match regardless of list order.

        '100 g pack date info' at 1.4mm comes FIRST but is NOT a substring of
        '100 g' (longer string can't be substr of shorter). '100 g' at 2.6mm
        is the only valid match and must be chosen.
        """
        resp = _run(FIXTURE_SPICE_AMBIGUOUS_FONT_MEASUREMENT)
        qty_entry = next(
            (s for s in resp.synthesized_fields_summary if s["field_type"] == "net_quantity"),
            None,
        )
        assert qty_entry is not None, "net_quantity must be synthesized from tags"
        assert qty_entry["font_height_mm"] == pytest.approx(2.6, abs=0.01), (
            f"Exact match '100 g' at 2.6mm must be chosen. "
            f"Got: {qty_entry['font_height_mm']} (first-list-entry 1.4mm would be wrong)"
        )

    def test_spice_font_height_above_threshold_no_violation(self):
        """2.6mm > 1.0mm printed threshold — FONT_SIZE_001 must not fire."""
        resp = _run(FIXTURE_SPICE_AMBIGUOUS_FONT_MEASUREMENT)
        font_violations = [v for v in resp.violations if v.rule_id == "FONT_SIZE_001"]
        assert font_violations == [], (
            f"2.6mm is above the 1.0mm printed threshold. "
            f"FONT_SIZE_001 must not fire. Got: {font_violations}"
        )

    # ------------------------------------------------------------------
    # CANDLE-TAGS-006: no measurement match → null → rule skipped
    # ------------------------------------------------------------------

    def test_candle_null_font_when_no_measurement_overlap(self):
        """No text_measurements entry overlaps with '150 g' → font_height_mm is None."""
        resp = _run(FIXTURE_CANDLE_NO_MEASUREMENT_MATCH)
        qty_entry = next(
            (s for s in resp.synthesized_fields_summary if s["field_type"] == "net_quantity"),
            None,
        )
        assert qty_entry is not None
        assert qty_entry["font_height_mm"] is None, (
            "No measurement overlap with '150 g' — font_height_mm must be None"
        )

    def test_candle_font_size_rule_skipped_when_null_font(self):
        """When font_height_mm is None, FONT_SIZE_001 must be silently skipped."""
        resp = _run(FIXTURE_CANDLE_NO_MEASUREMENT_MATCH)
        font_violations = [v for v in resp.violations if v.rule_id == "FONT_SIZE_001"]
        assert font_violations == [], (
            "font_height_mm=None — FONT_SIZE_001 must be skipped, not defaulted to pass or fail"
        )

    def test_candle_is_compliant_despite_null_font(self):
        """Null font height doesn't make the product non-compliant; engine must be compliant."""
        resp = _run(FIXTURE_CANDLE_NO_MEASUREMENT_MATCH)
        assert resp.overall_status == "compliant", (
            f"Null font should not trigger non-compliance. "
            f"Status: {resp.overall_status}, violations: {[v.rule_id for v in resp.violations]}"
        )

    # ------------------------------------------------------------------
    # COOKINGOIL-TAGS-001: imported + bonus-pack + food advisory
    # ------------------------------------------------------------------

    def test_cookingoil_importer_synthesized_from_tags(self):
        """IMPORTER_NAME/ADDRESS tags → 'importer' field synthesized (priority ①)."""
        resp = _run(FIXTURE_COOKINGOIL_IMPORTED_BONUS_PACK)
        synthesized_types = {s["field_type"] for s in resp.synthesized_fields_summary}
        assert "importer" in synthesized_types, (
            "IMPORTER_NAME + IMPORTER_ADDRESS in tags must produce a synthesized importer field"
        )

    def test_cookingoil_no_manufacturer_violation(self):
        """Importer declared via tags → MANUFACTURER_001 must not fire."""
        resp = _run(FIXTURE_COOKINGOIL_IMPORTED_BONUS_PACK)
        mfr_v = [v for v in resp.violations if v.rule_id == "MANUFACTURER_001"]
        assert mfr_v == [], (
            f"Importer synthesized from tags — MANUFACTURER_001 must not fire. Got: {mfr_v}"
        )

    def test_cookingoil_bonus_pack_quantity_valid(self):
        """'1 l + 100 ml FREE' bonus-pack notation must resolve to valid unit 'l'."""
        resp = _run(FIXTURE_COOKINGOIL_IMPORTED_BONUS_PACK)
        qty_v = [v for v in resp.violations if v.rule_id == "NET_QUANTITY_001"]
        assert qty_v == [], (
            f"'1 l + 100 ml FREE' is valid bonus-pack notation. "
            f"NET_QUANTITY_001 must not fire. Got: {qty_v}"
        )

    def test_cookingoil_font_height_resolved_from_measurement(self):
        """'1 l + 100 ml FREE' at 2.8mm resolved from text_measurements."""
        resp = _run(FIXTURE_COOKINGOIL_IMPORTED_BONUS_PACK)
        qty_entry = next(
            (s for s in resp.synthesized_fields_summary if s["field_type"] == "net_quantity"),
            None,
        )
        assert qty_entry is not None
        assert qty_entry["font_height_mm"] == pytest.approx(2.8, abs=0.01)

    def test_cookingoil_food_advisory_applied(self):
        """food category → mfg_date and net_quantity applicability notes have advisory."""
        resp = _run(FIXTURE_COOKINGOIL_IMPORTED_BONUS_PACK)
        advisory_notes = [
            n for n in resp.applicability_notes
            if n.severity_override == "advisory"
        ]
        assert len(advisory_notes) > 0, (
            "Food category must produce advisory applicability notes for mfg_date/net_quantity"
        )

    def test_cookingoil_country_of_origin_satisfied(self):
        """COUNTRY_OF_ORIGIN 'Greece' present → COUNTRY_ORIGIN_001 must not fire."""
        resp = _run(FIXTURE_COOKINGOIL_IMPORTED_BONUS_PACK)
        coo_v = [v for v in resp.violations if v.rule_id == "COUNTRY_ORIGIN_001"]
        assert coo_v == [], f"Country of origin satisfied. Got: {coo_v}"

    # ------------------------------------------------------------------
    # SHAMPOO-TAGS-004: e-commerce mfg_date regression
    # ------------------------------------------------------------------

    def test_shampoo_ecommerce_mfg_date_fires(self):
        """
        SHAMPOO-TAGS-004: e-commerce listing with no mfg_date tags.
        Rule 6(10) does NOT exempt physical label — MFG_DATE_001 must fire.
        Direct regression test for the Rule 6(10) correction.
        """
        resp = _run(FIXTURE_SHAMPOO_ECOMMERCE_MISSING_MFGDATE)
        mfg_v = [v for v in resp.violations if v.rule_id == "MFG_DATE_001"]
        assert len(mfg_v) > 0, (
            "mfg_date missing on physical label of e-commerce product — "
            "Rule 6(10) does not exempt physical label. MFG_DATE_001 must fire."
        )

    def test_shampoo_ecommerce_mfg_date_is_major(self):
        """mfg_date violation for 'other' category must be 'major' severity."""
        resp = _run(FIXTURE_SHAMPOO_ECOMMERCE_MISSING_MFGDATE)
        mfg_v = next((v for v in resp.violations if v.rule_id == "MFG_DATE_001"), None)
        assert mfg_v is not None
        assert mfg_v.severity == "major", (
            f"'other' category mfg_date violation must be major, got '{mfg_v.severity}'"
        )
