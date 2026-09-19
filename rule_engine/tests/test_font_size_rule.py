"""
tests/test_font_size_rule.py

Unit tests for FontSizeRule — Rule 7.
Tests cover: printed vs embossed thresholds, multi-field violations,
fields without font measurements, and check_all() behaviour.
"""

from __future__ import annotations

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from schemas import ExtractedField, ValidationRequest, BoundingBox
from rules.font_size_rule import FontSizeRule
from tests.test_fixtures import (
    FIXTURE_FONT_FAIL_PRINTED,
    FIXTURE_FONT_PASS_PRINTED,
    FIXTURE_FONT_FAIL_EMBOSSED,
    FIXTURE_FONT_PASS_PRINTED_1_5,
    FIXTURE_COMPLIANT_RETAIL,
)


def _make_field(
    field_type: str,
    font_height_mm: float | None = 1.5,
    confidence: float = 0.95,
    panel: str = "principal_display_panel",
) -> ExtractedField:
    return ExtractedField(
        field_type=field_type,
        text=f"Sample text for {field_type}",
        confidence=confidence,
        bbox=BoundingBox(x=0, y=0, w=100, h=20),
        font_height_mm=font_height_mm,
        panel_location=panel,
    )


def _make_request(fixture: dict) -> ValidationRequest:
    return ValidationRequest(**fixture)


RULE = FontSizeRule()


class TestFontSizePrintedPackage:
    """Tests for printed packaging (threshold = 1.0mm)."""

    def test_font_fail_0_8mm_printed(self):
        """0.8mm on printed — below 1.0mm threshold — violation."""
        req = _make_request(FIXTURE_FONT_FAIL_PRINTED)
        fields = [f for f in req.extracted_fields if f.confidence >= 0.75]
        violations = RULE.check_all(fields, req)
        failing_types = [v.field_type for v in violations]
        assert "net_quantity" in failing_types, f"Expected net_quantity to fail, got {failing_types}"

    def test_font_pass_1_5mm_printed(self):
        """1.5mm on printed — above 1.0mm threshold — no violation."""
        req = _make_request(FIXTURE_FONT_PASS_PRINTED)
        fields = [f for f in req.extracted_fields if f.confidence >= 0.75]
        violations = RULE.check_all(fields, req)
        assert violations == [], f"Expected no violations, got: {violations}"

    def test_font_pass_1_5mm_printed_fixture16(self):
        """1.5mm on printed (fixture 16) — above 1.0mm threshold — no violation."""
        req = _make_request(FIXTURE_FONT_PASS_PRINTED_1_5)
        fields = [f for f in req.extracted_fields if f.confidence >= 0.75]
        violations = RULE.check_all(fields, req)
        assert violations == [], f"Expected no violations, got: {violations}"

    def test_font_exactly_at_threshold_printed(self):
        """1.0mm exactly on printed — exactly meets threshold — no violation."""
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        fields = [_make_field("net_quantity", font_height_mm=1.0)]
        violations = RULE.check_all(fields, req)
        assert violations == [], "1.0mm should pass for printed (>= threshold)"

    def test_font_just_below_threshold_printed(self):
        """0.99mm on printed — just below 1.0mm — violation."""
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        fields = [_make_field("net_quantity", font_height_mm=0.99)]
        violations = RULE.check_all(fields, req)
        assert len(violations) == 1
        assert violations[0].clause == "Rule 7"


class TestFontSizeEmbossedPackage:
    """Tests for embossed/moulded/blown packaging (threshold = 2.0mm)."""

    def test_font_fail_1_5mm_embossed(self):
        """1.5mm on embossed — below 2.0mm threshold — violation."""
        req = _make_request(FIXTURE_FONT_FAIL_EMBOSSED)
        fields = [f for f in req.extracted_fields if f.confidence >= 0.75]
        violations = RULE.check_all(fields, req)
        failing_types = [v.field_type for v in violations]
        assert "net_quantity" in failing_types

    def test_font_pass_2_0mm_embossed(self):
        """2.0mm exactly on embossed — exactly meets threshold — no violation."""
        req = _make_request(FIXTURE_FONT_FAIL_EMBOSSED)
        # Override net_quantity font to 2.0mm (exact threshold)
        fields = [
            ExtractedField(
                field_type=f.field_type,
                text=f.text,
                confidence=f.confidence,
                bbox=f.bbox,
                font_height_mm=2.0 if f.field_type == "net_quantity" else f.font_height_mm,
                panel_location=f.panel_location,
            )
            for f in req.extracted_fields
            if f.confidence >= 0.75
        ]
        violations = RULE.check_all(fields, req)
        assert violations == [], "2.0mm should pass for embossed (>= threshold)"

    def test_font_fail_1_0mm_embossed(self):
        """1.0mm net_quantity on embossed (passes printed, fails embossed) — violation.

        Note: Rule 7's numeric mm threshold applies ONLY to net_quantity numerals.
        Other fields (mrp, manufacturer, etc.) are governed by Rule 6's qualitative
        'legible' standard — no mm figure is attached.
        """
        req = _make_request(FIXTURE_FONT_FAIL_EMBOSSED)
        fields = [_make_field("net_quantity", font_height_mm=1.0)]  # net_quantity only
        violations = RULE.check_all(fields, req)
        assert len(violations) == 1
        assert "2.0" in violations[0].description


class TestFontSizeMultipleViolations:
    """Tests for multiple font violations in one request."""

    def test_multiple_fields_failing(self):
        """Only net_quantity is checked by Rule 7 — mrp is NOT in scope.

        Rule 7 prescribes minimum numeral height ONLY for the net quantity
        declaration. Even if mrp has a small font, Rule 7 does not cover it
        (it falls under Rule 6's qualitative 'legible' standard).
        So: net_quantity (0.8mm) fails, mrp (0.7mm) is outside Rule 7 scope.
        Expected: 1 violation, not 2.
        """
        req = _make_request(FIXTURE_FONT_FAIL_PRINTED)
        fields = [
            _make_field("net_quantity", font_height_mm=0.8),  # FAIL — Rule 7 applies
            _make_field("mrp", font_height_mm=0.7),           # outside Rule 7 scope
        ]
        violations = RULE.check_all(fields, req)
        assert len(violations) == 1, (
            f"Only net_quantity falls under Rule 7 — mrp is not in scope. "
            f"Expected 1 violation, got {len(violations)}: {[v.field_type for v in violations]}"
        )
        assert violations[0].field_type == "net_quantity"

    def test_check_returns_first_only(self):
        """check() (singular) returns only first violation."""
        req = _make_request(FIXTURE_FONT_FAIL_PRINTED)
        fields = [
            _make_field("net_quantity", font_height_mm=0.8),
            _make_field("mrp", font_height_mm=0.7),
        ]
        result = RULE.check(fields, req)
        assert result is not None  # net_quantity violation only


class TestFontSizeMissingMeasurement:
    """Fields without font_height_mm are skipped — not flagged."""

    def test_no_font_measurement_skipped(self):
        """Field with font_height_mm=None — not assessed, no violation."""
        req = _make_request(FIXTURE_COMPLIANT_RETAIL)
        fields = [_make_field("net_quantity", font_height_mm=None)]
        violations = RULE.check_all(fields, req)
        assert violations == [], "Fields without font measurements should not be flagged"


class TestFontSizeRuleMetadata:
    def test_rule_id(self):
        assert RULE.rule_id == "FONT_SIZE_001"

    def test_clause(self):
        assert RULE.clause == "Rule 7"

    def test_severity(self):
        assert RULE.severity == "minor"
